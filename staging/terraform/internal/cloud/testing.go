// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"testing"
	"time"

	tfe "github.com/hashicorp/go-tfe"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/auth"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/colorstring"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/terraform/internal/backend"
	"github.com/hashicorp/terraform/internal/command/jsonformat"
	"github.com/hashicorp/terraform/internal/configs"
	"github.com/hashicorp/terraform/internal/configs/configschema"
	"github.com/hashicorp/terraform/internal/httpclient"
	"github.com/hashicorp/terraform/internal/providers"
	"github.com/hashicorp/terraform/internal/terraform"
	"github.com/hashicorp/terraform/internal/tfdiags"
	"github.com/hashicorp/terraform/version"

	backendLocal "github.com/hashicorp/terraform/internal/backend/local"
)

const (
	testCred = "test-auth-token"
)

var (
	tfeHost  = svchost.Hostname(defaultHostname)
	credsSrc = auth.StaticCredentialsSource(map[svchost.Hostname]map[string]interface{}{
		tfeHost: {"token": testCred},
	})
	testBackendSingleWorkspaceName = "app-prod"
	defaultTFCPing                 = map[string]func(http.ResponseWriter, *http.Request){
		"/api/v2/ping": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("TFP-API-Version", "2.5")
			w.Header().Set("TFP-AppName", "Terraform Cloud")
		},
	}
)

// mockInput is a mock implementation of terraform.UIInput.
type mockInput struct {
	answers map[string]string
}

func (m *mockInput) Input(ctx context.Context, opts *terraform.InputOpts) (string, error) {
	v, ok := m.answers[opts.Id]
	if !ok {
		return "", fmt.Errorf("unexpected input request in test: %s", opts.Id)
	}
	if v == "wait-for-external-update" {
		select {
		case <-ctx.Done():
		case <-time.After(time.Minute):
		}
	}
	delete(m.answers, opts.Id)
	return v, nil
}

func testInput(t *testing.T, answers map[string]string) *mockInput {
	return &mockInput{answers: answers}
}

func testBackendWithName(t *testing.T) (*Cloud, func()) {
	obj := cty.ObjectVal(map[string]cty.Value{
		"hostname":     cty.NullVal(cty.String),
		"organization": cty.StringVal("hashicorp"),
		"token":        cty.NullVal(cty.String),
		"workspaces": cty.ObjectVal(map[string]cty.Value{
			"name": cty.StringVal(testBackendSingleWorkspaceName),
			"tags": cty.NullVal(cty.Set(cty.String)),
		}),
	})
	return testBackend(t, obj, defaultTFCPing)
}

func testBackendWithTags(t *testing.T) (*Cloud, func()) {
	obj := cty.ObjectVal(map[string]cty.Value{
		"hostname":     cty.NullVal(cty.String),
		"organization": cty.StringVal("hashicorp"),
		"token":        cty.NullVal(cty.String),
		"workspaces": cty.ObjectVal(map[string]cty.Value{
			"name": cty.NullVal(cty.String),
			"tags": cty.SetVal(
				[]cty.Value{
					cty.StringVal("billing"),
				},
			),
		}),
	})
	return testBackend(t, obj, nil)
}

func testBackendNoOperations(t *testing.T) (*Cloud, func()) {
	obj := cty.ObjectVal(map[string]cty.Value{
		"hostname":     cty.NullVal(cty.String),
		"organization": cty.StringVal("no-operations"),
		"token":        cty.NullVal(cty.String),
		"workspaces": cty.ObjectVal(map[string]cty.Value{
			"name": cty.StringVal(testBackendSingleWorkspaceName),
			"tags": cty.NullVal(cty.Set(cty.String)),
		}),
	})
	return testBackend(t, obj, nil)
}

func testBackendWithHandlers(t *testing.T, handlers map[string]func(http.ResponseWriter, *http.Request)) (*Cloud, func()) {
	obj := cty.ObjectVal(map[string]cty.Value{
		"hostname":     cty.NullVal(cty.String),
		"organization": cty.StringVal("hashicorp"),
		"token":        cty.NullVal(cty.String),
		"workspaces": cty.ObjectVal(map[string]cty.Value{
			"name": cty.StringVal(testBackendSingleWorkspaceName),
			"tags": cty.NullVal(cty.Set(cty.String)),
		}),
	})
	return testBackend(t, obj, handlers)
}

func testCloudState(t *testing.T) *State {
	b, bCleanup := testBackendWithName(t)
	defer bCleanup()

	raw, err := b.StateMgr(testBackendSingleWorkspaceName)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	return raw.(*State)
}

func testBackendWithOutputs(t *testing.T) (*Cloud, func()) {
	b, cleanup := testBackendWithName(t)

	// Get a new mock client to use for adding outputs
	mc := NewMockClient()

	mc.StateVersionOutputs.create("svo-abcd", &tfe.StateVersionOutput{
		ID:           "svo-abcd",
		Value:        "foobar",
		Sensitive:    true,
		Type:         "string",
		Name:         "sensitive_output",
		DetailedType: "string",
	})

	mc.StateVersionOutputs.create("svo-zyxw", &tfe.StateVersionOutput{
		ID:           "svo-zyxw",
		Value:        "bazqux",
		Type:         "string",
		Name:         "nonsensitive_output",
		DetailedType: "string",
	})

	var dt interface{}
	var val interface{}
	err := json.Unmarshal([]byte(`["object", {"foo":"string"}]`), &dt)
	if err != nil {
		t.Fatalf("could not unmarshal detailed type: %s", err)
	}
	err = json.Unmarshal([]byte(`{"foo":"bar"}`), &val)
	if err != nil {
		t.Fatalf("could not unmarshal value: %s", err)
	}
	mc.StateVersionOutputs.create("svo-efgh", &tfe.StateVersionOutput{
		ID:           "svo-efgh",
		Value:        val,
		Type:         "object",
		Name:         "object_output",
		DetailedType: dt,
	})

	err = json.Unmarshal([]byte(`["list", "bool"]`), &dt)
	if err != nil {
		t.Fatalf("could not unmarshal detailed type: %s", err)
	}
	err = json.Unmarshal([]byte(`[true, false, true, true]`), &val)
	if err != nil {
		t.Fatalf("could not unmarshal value: %s", err)
	}
	mc.StateVersionOutputs.create("svo-ijkl", &tfe.StateVersionOutput{
		ID:           "svo-ijkl",
		Value:        val,
		Type:         "array",
		Name:         "list_output",
		DetailedType: dt,
	})

	b.client.StateVersionOutputs = mc.StateVersionOutputs

	return b, cleanup
}

func testBackend(t *testing.T, obj cty.Value, handlers map[string]func(http.ResponseWriter, *http.Request)) (*Cloud, func()) {
	var s *httptest.Server
	if handlers != nil {
		s = testServerWithHandlers(handlers)
	} else {
		s = testServer(t)
	}
	b := New(testDisco(s))

	// Configure the backend so the client is created.
	newObj, valDiags := b.PrepareConfig(obj)
	if len(valDiags) != 0 {
		t.Fatalf("testBackend: backend.PrepareConfig() failed: %s", valDiags.ErrWithWarnings())
	}
	obj = newObj

	confDiags := b.Configure(obj)
	if len(confDiags) != 0 {
		t.Fatalf("testBackend: backend.Configure() failed: %s", confDiags.ErrWithWarnings())
	}

	// Get a new mock client.
	mc := NewMockClient()

	// Replace the services we use with our mock services.
	b.CLI = cli.NewMockUi()
	b.client.Applies = mc.Applies
	b.client.ConfigurationVersions = mc.ConfigurationVersions
	b.client.CostEstimates = mc.CostEstimates
	b.client.Organizations = mc.Organizations
	b.client.Plans = mc.Plans
	b.client.TaskStages = mc.TaskStages
	b.client.PolicySetOutcomes = mc.PolicySetOutcomes
	b.client.PolicyChecks = mc.PolicyChecks
	b.client.Runs = mc.Runs
	b.client.RunEvents = mc.RunEvents
	b.client.StateVersions = mc.StateVersions
	b.client.StateVersionOutputs = mc.StateVersionOutputs
	b.client.Variables = mc.Variables
	b.client.Workspaces = mc.Workspaces

	// Set local to a local test backend.
	b.local = testLocalBackend(t, b)
	b.input = true

	baseURL, err := url.Parse("https://app.terraform.io")
	if err != nil {
		t.Fatalf("testBackend: failed to parse base URL for client")
	}
	baseURL.Path = "/api/v2/"

	readRedactedPlan = func(ctx context.Context, baseURL url.URL, token, planID string) (*jsonformat.Plan, error) {
		return mc.RedactedPlans.Read(ctx, baseURL.Hostname(), token, planID)
	}

	ctx := context.Background()

	// Create the organization.
	_, err = b.client.Organizations.Create(ctx, tfe.OrganizationCreateOptions{
		Name: tfe.String(b.organization),
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	// Create the default workspace if required.
	if b.WorkspaceMapping.Name != "" {
		_, err = b.client.Workspaces.Create(ctx, b.organization, tfe.WorkspaceCreateOptions{
			Name: tfe.String(b.WorkspaceMapping.Name),
		})
		if err != nil {
			t.Fatalf("error: %v", err)
		}
	}

	return b, s.Close
}

// testUnconfiguredBackend is used for testing the configuration of the backend
// with the mock client
func testUnconfiguredBackend(t *testing.T) (*Cloud, func()) {
	s := testServer(t)
	b := New(testDisco(s))

	// Normally, the client is created during configuration, but the configuration uses the
	// client to read entitlements.
	var err error
	b.client, err = tfe.NewClient(&tfe.Config{
		Token: "fake-token",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Get a new mock client.
	mc := NewMockClient()

	// Replace the services we use with our mock services.
	b.CLI = cli.NewMockUi()
	b.client.Applies = mc.Applies
	b.client.ConfigurationVersions = mc.ConfigurationVersions
	b.client.CostEstimates = mc.CostEstimates
	b.client.Organizations = mc.Organizations
	b.client.Plans = mc.Plans
	b.client.PolicySetOutcomes = mc.PolicySetOutcomes
	b.client.PolicyChecks = mc.PolicyChecks
	b.client.Runs = mc.Runs
	b.client.RunEvents = mc.RunEvents
	b.client.StateVersions = mc.StateVersions
	b.client.Variables = mc.Variables
	b.client.Workspaces = mc.Workspaces

	baseURL, err := url.Parse("https://app.terraform.io")
	if err != nil {
		t.Fatalf("testBackend: failed to parse base URL for client")
	}
	baseURL.Path = "/api/v2/"

	readRedactedPlan = func(ctx context.Context, baseURL url.URL, token, planID string) (*jsonformat.Plan, error) {
		return mc.RedactedPlans.Read(ctx, baseURL.Hostname(), token, planID)
	}

	// Set local to a local test backend.
	b.local = testLocalBackend(t, b)

	return b, s.Close
}

func testLocalBackend(t *testing.T, cloud *Cloud) backend.Enhanced {
	b := backendLocal.NewWithBackend(cloud)

	// Add a test provider to the local backend.
	p := backendLocal.TestLocalProvider(t, b, "null", &terraform.ProviderSchema{
		ResourceTypes: map[string]*configschema.Block{
			"null_resource": {
				Attributes: map[string]*configschema.Attribute{
					"id": {Type: cty.String, Computed: true},
				},
			},
		},
	})
	p.ApplyResourceChangeResponse = &providers.ApplyResourceChangeResponse{NewState: cty.ObjectVal(map[string]cty.Value{
		"id": cty.StringVal("yes"),
	})}

	return b
}

// testServer returns a started *httptest.Server used for local testing with the default set of
// request handlers.
func testServer(t *testing.T) *httptest.Server {
	return testServerWithHandlers(testDefaultRequestHandlers)
}

// testServerWithHandlers returns a started *httptest.Server with the given set of request handlers
// overriding any default request handlers (testDefaultRequestHandlers).
func testServerWithHandlers(handlers map[string]func(http.ResponseWriter, *http.Request)) *httptest.Server {
	mux := http.NewServeMux()
	for route, handler := range handlers {
		mux.HandleFunc(route, handler)
	}
	for route, handler := range testDefaultRequestHandlers {
		if handlers[route] == nil {
			mux.HandleFunc(route, handler)
		}
	}

	return httptest.NewServer(mux)
}

// testDefaultRequestHandlers is a map of request handlers intended to be used in a request
// multiplexer for a test server. A caller may use testServerWithHandlers to start a server with
// this base set of routes, and override a particular route for whatever edge case is being tested.
var testDefaultRequestHandlers = map[string]func(http.ResponseWriter, *http.Request){
	// Respond to service discovery calls.
	"/well-known/terraform.json": func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{
  "tfe.v2": "/api/v2/",
}`)
	},

	// Respond to service version constraints calls.
	"/v1/versions/": func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, fmt.Sprintf(`{
  "service": "%s",
  "product": "terraform",
  "minimum": "0.1.0",
  "maximum": "10.0.0"
}`, path.Base(r.URL.Path)))
	},

	// Respond to pings to get the API version header.
	"/api/v2/ping": func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("TFP-API-Version", "2.5")
	},

	// Respond to the initial query to read the hashicorp org entitlements.
	"/api/v2/organizations/hashicorp/entitlement-set": func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		io.WriteString(w, `{
  "data": {
    "id": "org-GExadygjSbKP8hsY",
    "type": "entitlement-sets",
    "attributes": {
      "operations": true,
      "private-module-registry": true,
      "sentinel": true,
      "state-storage": true,
      "teams": true,
      "vcs-integrations": true
    }
  }
}`)
	},

	// Respond to the initial query to read the no-operations org entitlements.
	"/api/v2/organizations/no-operations/entitlement-set": func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		io.WriteString(w, `{
  "data": {
    "id": "org-ufxa3y8jSbKP8hsT",
    "type": "entitlement-sets",
    "attributes": {
      "operations": false,
      "private-module-registry": true,
      "sentinel": true,
      "state-storage": true,
      "teams": true,
      "vcs-integrations": true
    }
  }
}`)
	},

	// All tests that are assumed to pass will use the hashicorp organization,
	// so for all other organization requests we will return a 404.
	"/api/v2/organizations/": func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		io.WriteString(w, `{
  "errors": [
    {
      "status": "404",
      "title": "not found"
    }
  ]
}`)
	},
}

func mockColorize() *colorstring.Colorize {
	colors := make(map[string]string)
	for k, v := range colorstring.DefaultColors {
		colors[k] = v
	}
	colors["purple"] = "38;5;57"

	return &colorstring.Colorize{
		Colors:  colors,
		Disable: false,
		Reset:   true,
	}
}

func mockSROWorkspace(t *testing.T, b *Cloud, workspaceName string) {
	_, err := b.client.Workspaces.Update(context.Background(), "hashicorp", workspaceName, tfe.WorkspaceUpdateOptions{
		StructuredRunOutputEnabled: tfe.Bool(true),
		TerraformVersion:           tfe.String("1.4.0"),
	})
	if err != nil {
		t.Fatalf("Error enabling SRO on workspace %s: %v", workspaceName, err)
	}
}

// testDisco returns a *disco.Disco mapping app.terraform.io and
// localhost to a local test server.
func testDisco(s *httptest.Server) *disco.Disco {
	services := map[string]interface{}{
		"tfe.v2": fmt.Sprintf("%s/api/v2/", s.URL),
	}
	d := disco.NewWithCredentialsSource(credsSrc)
	d.SetUserAgent(httpclient.TerraformUserAgent(version.String()))

	d.ForceHostServices(svchost.Hostname(defaultHostname), services)
	d.ForceHostServices(svchost.Hostname("localhost"), services)
	d.ForceHostServices(svchost.Hostname("nontfe.local"), nil)
	return d
}

type unparsedVariableValue struct {
	value  string
	source terraform.ValueSourceType
}

func (v *unparsedVariableValue) ParseVariableValue(mode configs.VariableParsingMode) (*terraform.InputValue, tfdiags.Diagnostics) {
	return &terraform.InputValue{
		Value:      cty.StringVal(v.value),
		SourceType: v.source,
	}, tfdiags.Diagnostics{}
}

// testVariable returns a backend.UnparsedVariableValue used for testing.
func testVariables(s terraform.ValueSourceType, vs ...string) map[string]backend.UnparsedVariableValue {
	vars := make(map[string]backend.UnparsedVariableValue, len(vs))
	for _, v := range vs {
		vars[v] = &unparsedVariableValue{
			value:  v,
			source: s,
		}
	}
	return vars
}

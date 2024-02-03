// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/hashicorp/terraform/internal/backend"
	"github.com/hashicorp/terraform/internal/legacy/helper/schema"
	"github.com/hashicorp/terraform/internal/logging"
	"github.com/hashicorp/terraform/internal/states/remote"
	"github.com/hashicorp/terraform/internal/states/statemgr"
)

func New() backend.Backend {
	s := &schema.Backend{
		Schema: map[string]*schema.Schema{
			"address": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_ADDRESS", nil),
				Description: "The address of the REST endpoint",
			},
			"update_method": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_UPDATE_METHOD", "POST"),
				Description: "HTTP method to use when updating state",
			},
			"lock_address": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_LOCK_ADDRESS", nil),
				Description: "The address of the lock REST endpoint",
			},
			"unlock_address": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_UNLOCK_ADDRESS", nil),
				Description: "The address of the unlock REST endpoint",
			},
			"lock_method": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_LOCK_METHOD", "LOCK"),
				Description: "The HTTP method to use when locking",
			},
			"unlock_method": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_UNLOCK_METHOD", "UNLOCK"),
				Description: "The HTTP method to use when unlocking",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_USERNAME", nil),
				Description: "The username for HTTP basic authentication",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_PASSWORD", nil),
				Description: "The password for HTTP basic authentication",
			},
			"skip_cert_verification": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to skip TLS verification.",
			},
			"retry_max": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_RETRY_MAX", 2),
				Description: "The number of HTTP request retries.",
			},
			"retry_wait_min": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_RETRY_WAIT_MIN", 1),
				Description: "The minimum time in seconds to wait between HTTP request attempts.",
			},
			"retry_wait_max": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_RETRY_WAIT_MAX", 30),
				Description: "The maximum time in seconds to wait between HTTP request attempts.",
			},
			"client_ca_certificate_pem": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_CLIENT_CA_CERTIFICATE_PEM", ""),
				Description: "A PEM-encoded CA certificate chain used by the client to verify server certificates during TLS authentication.",
			},
			"client_certificate_pem": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_CLIENT_CERTIFICATE_PEM", ""),
				Description: "A PEM-encoded certificate used by the server to verify the client during mutual TLS (mTLS) authentication.",
			},
			"client_private_key_pem": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("TF_HTTP_CLIENT_PRIVATE_KEY_PEM", ""),
				Description: "A PEM-encoded private key, required if client_certificate_pem is specified.",
			},
		},
	}

	b := &Backend{Backend: s}
	b.Backend.ConfigureFunc = b.configure
	return b
}

type Backend struct {
	*schema.Backend

	client *httpClient
}

// configureTLS configures TLS when needed; if there are no conditions requiring TLS, no change is made.
func (b *Backend) configureTLS(client *retryablehttp.Client, data *schema.ResourceData) error {
	// If there are no conditions needing to configure TLS, leave the client untouched
	skipCertVerification := data.Get("skip_cert_verification").(bool)
	clientCACertificatePem := data.Get("client_ca_certificate_pem").(string)
	clientCertificatePem := data.Get("client_certificate_pem").(string)
	clientPrivateKeyPem := data.Get("client_private_key_pem").(string)
	if !skipCertVerification && clientCACertificatePem == "" && clientCertificatePem == "" && clientPrivateKeyPem == "" {
		return nil
	}
	if clientCertificatePem != "" && clientPrivateKeyPem == "" {
		return fmt.Errorf("client_certificate_pem is set but client_private_key_pem is not")
	}
	if clientPrivateKeyPem != "" && clientCertificatePem == "" {
		return fmt.Errorf("client_private_key_pem is set but client_certificate_pem is not")
	}

	// TLS configuration is needed; create an object and configure it
	var tlsConfig tls.Config
	client.HTTPClient.Transport.(*http.Transport).TLSClientConfig = &tlsConfig

	if skipCertVerification {
		// ignores TLS verification
		tlsConfig.InsecureSkipVerify = true
	}
	if clientCACertificatePem != "" {
		// trust servers based on a CA
		tlsConfig.RootCAs = x509.NewCertPool()
		if !tlsConfig.RootCAs.AppendCertsFromPEM([]byte(clientCACertificatePem)) {
			return errors.New("failed to append certs")
		}
	}
	if clientCertificatePem != "" && clientPrivateKeyPem != "" {
		// attach a client certificate to the TLS handshake (aka mTLS)
		certificate, err := tls.X509KeyPair([]byte(clientCertificatePem), []byte(clientPrivateKeyPem))
		if err != nil {
			return fmt.Errorf("cannot load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}
	}

	return nil
}

func (b *Backend) configure(ctx context.Context) error {
	data := schema.FromContextBackendConfig(ctx)

	address := data.Get("address").(string)
	updateURL, err := url.Parse(address)
	if err != nil {
		return fmt.Errorf("failed to parse address URL: %s", err)
	}
	if updateURL.Scheme != "http" && updateURL.Scheme != "https" {
		return fmt.Errorf("address must be HTTP or HTTPS")
	}

	updateMethod := data.Get("update_method").(string)

	var lockURL *url.URL
	if v, ok := data.GetOk("lock_address"); ok && v.(string) != "" {
		var err error
		lockURL, err = url.Parse(v.(string))
		if err != nil {
			return fmt.Errorf("failed to parse lockAddress URL: %s", err)
		}
		if lockURL.Scheme != "http" && lockURL.Scheme != "https" {
			return fmt.Errorf("lockAddress must be HTTP or HTTPS")
		}
	}

	lockMethod := data.Get("lock_method").(string)

	var unlockURL *url.URL
	if v, ok := data.GetOk("unlock_address"); ok && v.(string) != "" {
		var err error
		unlockURL, err = url.Parse(v.(string))
		if err != nil {
			return fmt.Errorf("failed to parse unlockAddress URL: %s", err)
		}
		if unlockURL.Scheme != "http" && unlockURL.Scheme != "https" {
			return fmt.Errorf("unlockAddress must be HTTP or HTTPS")
		}
	}

	unlockMethod := data.Get("unlock_method").(string)

	rClient := retryablehttp.NewClient()
	rClient.RetryMax = data.Get("retry_max").(int)
	rClient.RetryWaitMin = time.Duration(data.Get("retry_wait_min").(int)) * time.Second
	rClient.RetryWaitMax = time.Duration(data.Get("retry_wait_max").(int)) * time.Second
	rClient.Logger = log.New(logging.LogOutput(), "", log.Flags())
	if err = b.configureTLS(rClient, data); err != nil {
		return err
	}

	b.client = &httpClient{
		URL:          updateURL,
		UpdateMethod: updateMethod,

		LockURL:      lockURL,
		LockMethod:   lockMethod,
		UnlockURL:    unlockURL,
		UnlockMethod: unlockMethod,

		Username: data.Get("username").(string),
		Password: data.Get("password").(string),

		// accessible only for testing use
		Client: rClient,
	}
	return nil
}

func (b *Backend) StateMgr(name string) (statemgr.Full, error) {
	if name != backend.DefaultStateName {
		return nil, backend.ErrWorkspacesNotSupported
	}

	return &remote.State{Client: b.client}, nil
}

func (b *Backend) Workspaces() ([]string, error) {
	return nil, backend.ErrWorkspacesNotSupported
}

func (b *Backend) DeleteWorkspace(string, bool) error {
	return backend.ErrWorkspacesNotSupported
}

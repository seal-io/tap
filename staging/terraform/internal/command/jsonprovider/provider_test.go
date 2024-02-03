// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package jsonprovider

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/terraform/internal/configs/configschema"
	"github.com/hashicorp/terraform/internal/terraform"
)

func TestMarshalProvider(t *testing.T) {
	tests := []struct {
		Input *terraform.ProviderSchema
		Want  *Provider
	}{
		{
			nil,
			&Provider{},
		},
		{
			testProvider(),
			&Provider{
				Provider: &Schema{
					Block: &Block{
						Attributes: map[string]*Attribute{
							"region": {
								AttributeType:   json.RawMessage(`"string"`),
								Required:        true,
								DescriptionKind: "plain",
							},
						},
						DescriptionKind: "plain",
					},
				},
				ResourceSchemas: map[string]*Schema{
					"test_instance": {
						Version: 42,
						Block: &Block{
							Attributes: map[string]*Attribute{
								"id": {
									AttributeType:   json.RawMessage(`"string"`),
									Optional:        true,
									Computed:        true,
									DescriptionKind: "plain",
								},
								"ami": {
									AttributeType:   json.RawMessage(`"string"`),
									Optional:        true,
									DescriptionKind: "plain",
								},
								"volumes": {
									AttributeNestedType: &NestedType{
										NestingMode: "list",
										Attributes: map[string]*Attribute{
											"size": {
												AttributeType:   json.RawMessage(`"string"`),
												Required:        true,
												DescriptionKind: "plain",
											},
											"mount_point": {
												AttributeType:   json.RawMessage(`"string"`),
												Required:        true,
												DescriptionKind: "plain",
											},
										},
									},
									Optional:        true,
									DescriptionKind: "plain",
								},
							},
							BlockTypes: map[string]*BlockType{
								"network_interface": {
									Block: &Block{
										Attributes: map[string]*Attribute{
											"device_index": {
												AttributeType:   json.RawMessage(`"string"`),
												Optional:        true,
												DescriptionKind: "plain",
											},
											"description": {
												AttributeType:   json.RawMessage(`"string"`),
												Optional:        true,
												DescriptionKind: "plain",
											},
										},
										DescriptionKind: "plain",
									},
									NestingMode: "list",
								},
							},
							DescriptionKind: "plain",
						},
					},
				},
				DataSourceSchemas: map[string]*Schema{
					"test_data_source": {
						Version: 3,
						Block: &Block{
							Attributes: map[string]*Attribute{
								"id": {
									AttributeType:   json.RawMessage(`"string"`),
									Optional:        true,
									Computed:        true,
									DescriptionKind: "plain",
								},
								"ami": {
									AttributeType:   json.RawMessage(`"string"`),
									Optional:        true,
									DescriptionKind: "plain",
								},
							},
							BlockTypes: map[string]*BlockType{
								"network_interface": {
									Block: &Block{
										Attributes: map[string]*Attribute{
											"device_index": {
												AttributeType:   json.RawMessage(`"string"`),
												Optional:        true,
												DescriptionKind: "plain",
											},
											"description": {
												AttributeType:   json.RawMessage(`"string"`),
												Optional:        true,
												DescriptionKind: "plain",
											},
										},
										DescriptionKind: "plain",
									},
									NestingMode: "list",
								},
							},
							DescriptionKind: "plain",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		got := marshalProvider(test.Input)
		if !cmp.Equal(got, test.Want) {
			t.Fatalf("wrong result:\n %v\n", cmp.Diff(got, test.Want))
		}
	}
}

func testProvider() *terraform.ProviderSchema {
	return &terraform.ProviderSchema{
		Provider: &configschema.Block{
			Attributes: map[string]*configschema.Attribute{
				"region": {Type: cty.String, Required: true},
			},
		},
		ResourceTypes: map[string]*configschema.Block{
			"test_instance": {
				Attributes: map[string]*configschema.Attribute{
					"id":  {Type: cty.String, Optional: true, Computed: true},
					"ami": {Type: cty.String, Optional: true},
					"volumes": {
						Optional: true,
						NestedType: &configschema.Object{
							Nesting: configschema.NestingList,
							Attributes: map[string]*configschema.Attribute{
								"size":        {Type: cty.String, Required: true},
								"mount_point": {Type: cty.String, Required: true},
							},
						},
					},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"network_interface": {
						Nesting: configschema.NestingList,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"device_index": {Type: cty.String, Optional: true},
								"description":  {Type: cty.String, Optional: true},
							},
						},
					},
				},
			},
		},
		DataSources: map[string]*configschema.Block{
			"test_data_source": {
				Attributes: map[string]*configschema.Attribute{
					"id":  {Type: cty.String, Optional: true, Computed: true},
					"ami": {Type: cty.String, Optional: true},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"network_interface": {
						Nesting: configschema.NestingList,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"device_index": {Type: cty.String, Optional: true},
								"description":  {Type: cty.String, Optional: true},
							},
						},
					},
				},
			},
		},

		ResourceTypeSchemaVersions: map[string]uint64{
			"test_instance":    42,
			"test_data_source": 3,
		},
	}
}

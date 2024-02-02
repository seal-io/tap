// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/configs/hcl2shim"
	"github.com/hashicorp/terraform/legacy/helper/hashcode"
	"github.com/hashicorp/terraform/legacy/terraform"
)

func TestConfigFieldReader_impl(t *testing.T) {
	var _ FieldReader = new(ConfigFieldReader)
}

func TestConfigFieldReader(t *testing.T) {
	testFieldReader(t, func(s map[string]*Schema) FieldReader {
		return &ConfigFieldReader{
			Schema: s,

			Config: testConfig(t, map[string]any{
				"bool":   true,
				"float":  3.1415,
				"int":    42,
				"string": "string",

				"list": []any{"foo", "bar"},

				"listInt": []any{21, 42},

				"map": map[string]any{
					"foo": "bar",
					"bar": "baz",
				},
				"mapInt": map[string]any{
					"one": "1",
					"two": "2",
				},
				"mapIntNestedSchema": map[string]any{
					"one": "1",
					"two": "2",
				},
				"mapFloat": map[string]any{
					"oneDotTwo": "1.2",
				},
				"mapBool": map[string]any{
					"True":  "true",
					"False": "false",
				},

				"set": []any{10, 50},
				"setDeep": []any{
					map[string]any{
						"index": 10,
						"value": "foo",
					},
					map[string]any{
						"index": 50,
						"value": "bar",
					},
				},
			}),
		}
	})
}

// This contains custom table tests for our ConfigFieldReader
func TestConfigFieldReader_custom(t *testing.T) {
	schema := map[string]*Schema{
		"bool": &Schema{
			Type: TypeBool,
		},
	}

	cases := map[string]struct {
		Addr   []string
		Result FieldReadResult
		Config *terraform.ResourceConfig
		Err    bool
	}{
		"basic": {
			[]string{"bool"},
			FieldReadResult{
				Value:  true,
				Exists: true,
			},
			testConfig(t, map[string]any{
				"bool": true,
			}),
			false,
		},

		"computed": {
			[]string{"bool"},
			FieldReadResult{
				Exists:   true,
				Computed: true,
			},
			testConfig(t, map[string]any{
				"bool": hcl2shim.UnknownVariableValue,
			}),
			false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &ConfigFieldReader{
				Schema: schema,
				Config: tc.Config,
			}
			out, err := r.ReadField(tc.Addr)
			if err != nil != tc.Err {
				t.Fatalf("%s: err: %s", name, err)
			}
			if s, ok := out.Value.(*Set); ok {
				// If it is a set, convert to a list so its more easily checked.
				out.Value = s.List()
			}
			if !reflect.DeepEqual(tc.Result, out) {
				t.Fatalf("%s: bad: %#v", name, out)
			}
		})
	}
}

func TestConfigFieldReader_DefaultHandling(t *testing.T) {
	schema := map[string]*Schema{
		"strWithDefault": &Schema{
			Type:    TypeString,
			Default: "ImADefault",
		},
		"strWithDefaultFunc": &Schema{
			Type: TypeString,
			DefaultFunc: func() (any, error) {
				return "FuncDefault", nil
			},
		},
	}

	cases := map[string]struct {
		Addr   []string
		Result FieldReadResult
		Config *terraform.ResourceConfig
		Err    bool
	}{
		"gets default value when no config set": {
			[]string{"strWithDefault"},
			FieldReadResult{
				Value:    "ImADefault",
				Exists:   true,
				Computed: false,
			},
			testConfig(t, map[string]any{}),
			false,
		},
		"config overrides default value": {
			[]string{"strWithDefault"},
			FieldReadResult{
				Value:    "fromConfig",
				Exists:   true,
				Computed: false,
			},
			testConfig(t, map[string]any{
				"strWithDefault": "fromConfig",
			}),
			false,
		},
		"gets default from function when no config set": {
			[]string{"strWithDefaultFunc"},
			FieldReadResult{
				Value:    "FuncDefault",
				Exists:   true,
				Computed: false,
			},
			testConfig(t, map[string]any{}),
			false,
		},
		"config overrides default function": {
			[]string{"strWithDefaultFunc"},
			FieldReadResult{
				Value:    "fromConfig",
				Exists:   true,
				Computed: false,
			},
			testConfig(t, map[string]any{
				"strWithDefaultFunc": "fromConfig",
			}),
			false,
		},
	}

	for name, tc := range cases {
		r := &ConfigFieldReader{
			Schema: schema,
			Config: tc.Config,
		}
		out, err := r.ReadField(tc.Addr)
		if err != nil != tc.Err {
			t.Fatalf("%s: err: %s", name, err)
		}
		if s, ok := out.Value.(*Set); ok {
			// If it is a set, convert to a list so its more easily checked.
			out.Value = s.List()
		}
		if !reflect.DeepEqual(tc.Result, out) {
			t.Fatalf("%s: bad: %#v", name, out)
		}
	}
}

func TestConfigFieldReader_ComputedMap(t *testing.T) {
	schema := map[string]*Schema{
		"map": &Schema{
			Type:     TypeMap,
			Computed: true,
		},
		"listmap": &Schema{
			Type:     TypeMap,
			Computed: true,
			Elem:     TypeList,
		},
		"maplist": &Schema{
			Type:     TypeList,
			Computed: true,
			Elem:     TypeMap,
		},
	}

	cases := []struct {
		Name   string
		Addr   []string
		Result FieldReadResult
		Config *terraform.ResourceConfig
		Err    bool
	}{
		{
			"set, normal",
			[]string{"map"},
			FieldReadResult{
				Value: map[string]any{
					"foo": "bar",
				},
				Exists:   true,
				Computed: false,
			},
			testConfig(t, map[string]any{
				"map": map[string]any{
					"foo": "bar",
				},
			}),
			false,
		},

		{
			"computed element",
			[]string{"map"},
			FieldReadResult{
				Exists:   true,
				Computed: true,
			},
			testConfig(t, map[string]any{
				"map": map[string]any{
					"foo": hcl2shim.UnknownVariableValue,
				},
			}),
			false,
		},

		{
			"native map",
			[]string{"map"},
			FieldReadResult{
				Value: map[string]any{
					"bar": "baz",
					"baz": "bar",
				},
				Exists:   true,
				Computed: false,
			},
			testConfig(t, map[string]any{
				"map": map[string]any{
					"bar": "baz",
					"baz": "bar",
				},
			}),
			false,
		},

		{
			"map-from-list-of-maps",
			[]string{"maplist", "0"},
			FieldReadResult{
				Value: map[string]any{
					"key": "bar",
				},
				Exists:   true,
				Computed: false,
			},
			testConfig(t, map[string]any{
				"maplist": []any{
					map[string]any{
						"key": "bar",
					},
				},
			}),
			false,
		},

		{
			"value-from-list-of-maps",
			[]string{"maplist", "0", "key"},
			FieldReadResult{
				Value:    "bar",
				Exists:   true,
				Computed: false,
			},
			testConfig(t, map[string]any{
				"maplist": []any{
					map[string]any{
						"key": "bar",
					},
				},
			}),
			false,
		},

		{
			"list-from-map-of-lists",
			[]string{"listmap", "key"},
			FieldReadResult{
				Value:    []any{"bar"},
				Exists:   true,
				Computed: false,
			},
			testConfig(t, map[string]any{
				"listmap": map[string]any{
					"key": []any{
						"bar",
					},
				},
			}),
			false,
		},

		{
			"value-from-map-of-lists",
			[]string{"listmap", "key", "0"},
			FieldReadResult{
				Value:    "bar",
				Exists:   true,
				Computed: false,
			},
			testConfig(t, map[string]any{
				"listmap": map[string]any{
					"key": []any{
						"bar",
					},
				},
			}),
			false,
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.Name), func(t *testing.T) {
			r := &ConfigFieldReader{
				Schema: schema,
				Config: tc.Config,
			}
			out, err := r.ReadField(tc.Addr)
			if err != nil != tc.Err {
				t.Fatal(err)
			}
			if s, ok := out.Value.(*Set); ok {
				// If it is a set, convert to the raw map
				out.Value = s.m
				if len(s.m) == 0 {
					out.Value = nil
				}
			}
			if !reflect.DeepEqual(tc.Result, out) {
				t.Fatalf("\nexpected: %#v\ngot:      %#v", tc.Result, out)
			}
		})
	}
}

func TestConfigFieldReader_ComputedSet(t *testing.T) {
	schema := map[string]*Schema{
		"strSet": &Schema{
			Type: TypeSet,
			Elem: &Schema{Type: TypeString},
			Set:  HashString,
		},
	}

	cases := map[string]struct {
		Addr   []string
		Result FieldReadResult
		Config *terraform.ResourceConfig
		Err    bool
	}{
		"set, normal": {
			[]string{"strSet"},
			FieldReadResult{
				Value: map[string]any{
					"2356372769": "foo",
				},
				Exists:   true,
				Computed: false,
			},
			testConfig(t, map[string]any{
				"strSet": []any{"foo"},
			}),
			false,
		},

		"set, computed element": {
			[]string{"strSet"},
			FieldReadResult{
				Value:    nil,
				Exists:   true,
				Computed: true,
			},
			testConfig(t, map[string]any{
				"strSet": []any{hcl2shim.UnknownVariableValue},
			}),
			false,
		},
	}

	for name, tc := range cases {
		r := &ConfigFieldReader{
			Schema: schema,
			Config: tc.Config,
		}
		out, err := r.ReadField(tc.Addr)
		if err != nil != tc.Err {
			t.Fatalf("%s: err: %s", name, err)
		}
		if s, ok := out.Value.(*Set); ok {
			// If it is a set, convert to the raw map
			out.Value = s.m
			if len(s.m) == 0 {
				out.Value = nil
			}
		}
		if !reflect.DeepEqual(tc.Result, out) {
			t.Fatalf("%s: bad: %#v", name, out)
		}
	}
}

func TestConfigFieldReader_computedComplexSet(t *testing.T) {
	hashfunc := func(v any) int {
		var buf bytes.Buffer
		m := v.(map[string]any)
		buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
		buf.WriteString(fmt.Sprintf("%s-", m["vhd_uri"].(string)))
		return hashcode.String(buf.String())
	}

	schema := map[string]*Schema{
		"set": &Schema{
			Type: TypeSet,
			Elem: &Resource{
				Schema: map[string]*Schema{
					"name": {
						Type:     TypeString,
						Required: true,
					},

					"vhd_uri": {
						Type:     TypeString,
						Required: true,
					},
				},
			},
			Set: hashfunc,
		},
	}

	cases := map[string]struct {
		Addr   []string
		Result FieldReadResult
		Config *terraform.ResourceConfig
		Err    bool
	}{
		"set, normal": {
			[]string{"set"},
			FieldReadResult{
				Value: map[string]any{
					"532860136": map[string]any{
						"name":    "myosdisk1",
						"vhd_uri": "bar",
					},
				},
				Exists:   true,
				Computed: false,
			},
			testConfig(t, map[string]any{
				"set": []any{
					map[string]any{
						"name":    "myosdisk1",
						"vhd_uri": "bar",
					},
				},
			}),
			false,
		},
	}

	for name, tc := range cases {
		r := &ConfigFieldReader{
			Schema: schema,
			Config: tc.Config,
		}
		out, err := r.ReadField(tc.Addr)
		if err != nil != tc.Err {
			t.Fatalf("%s: err: %s", name, err)
		}
		if s, ok := out.Value.(*Set); ok {
			// If it is a set, convert to the raw map
			out.Value = s.m
			if len(s.m) == 0 {
				out.Value = nil
			}
		}
		if !reflect.DeepEqual(tc.Result, out) {
			t.Fatalf("%s: bad: %#v", name, out)
		}
	}
}

func testConfig(t *testing.T, raw map[string]any) *terraform.ResourceConfig {
	return terraform.NewResourceConfigRaw(raw)
}

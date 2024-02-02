// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"testing"
)

func TestSerializeForHash(t *testing.T) {
	type testCase struct {
		Schema   any
		Value    any
		Expected string
	}

	tests := []testCase{
		testCase{
			Schema: &Schema{
				Type: TypeInt,
			},
			Value:    0,
			Expected: "0;",
		},

		testCase{
			Schema: &Schema{
				Type: TypeInt,
			},
			Value:    200,
			Expected: "200;",
		},

		testCase{
			Schema: &Schema{
				Type: TypeBool,
			},
			Value:    true,
			Expected: "1;",
		},

		testCase{
			Schema: &Schema{
				Type: TypeBool,
			},
			Value:    false,
			Expected: "0;",
		},

		testCase{
			Schema: &Schema{
				Type: TypeFloat,
			},
			Value:    1.0,
			Expected: "1;",
		},

		testCase{
			Schema: &Schema{
				Type: TypeFloat,
			},
			Value:    1.54,
			Expected: "1.54;",
		},

		testCase{
			Schema: &Schema{
				Type: TypeFloat,
			},
			Value:    0.1,
			Expected: "0.1;",
		},

		testCase{
			Schema: &Schema{
				Type: TypeString,
			},
			Value:    "hello",
			Expected: "hello;",
		},

		testCase{
			Schema: &Schema{
				Type: TypeString,
			},
			Value:    "1",
			Expected: "1;",
		},

		testCase{
			Schema: &Schema{
				Type: TypeList,
				Elem: &Schema{
					Type: TypeString,
				},
			},
			Value:    []any{},
			Expected: "();",
		},

		testCase{
			Schema: &Schema{
				Type: TypeList,
				Elem: &Schema{
					Type: TypeString,
				},
			},
			Value:    []any{"hello", "world"},
			Expected: "(hello;world;);",
		},

		testCase{
			Schema: &Schema{
				Type: TypeList,
				Elem: &Resource{
					Schema: map[string]*Schema{
						"fo": &Schema{
							Type:     TypeString,
							Required: true,
						},
						"fum": &Schema{
							Type:     TypeString,
							Required: true,
						},
					},
				},
			},
			Value: []any{
				map[string]any{
					"fo": "bar",
				},
				map[string]any{
					"fo":  "baz",
					"fum": "boz",
				},
			},
			Expected: "(<fo:bar;fum:;>;<fo:baz;fum:boz;>;);",
		},

		testCase{
			Schema: &Schema{
				Type: TypeSet,
				Elem: &Schema{
					Type: TypeString,
				},
			},
			Value: NewSet(func(i any) int { return len(i.(string)) }, []any{
				"hello",
				"woo",
			}),
			Expected: "{woo;hello;};",
		},

		testCase{
			Schema: &Schema{
				Type: TypeMap,
				Elem: &Schema{
					Type: TypeString,
				},
			},
			Value: map[string]any{
				"foo": "bar",
				"baz": "foo",
			},
			Expected: "[baz:foo;foo:bar;];",
		},

		testCase{
			Schema: &Resource{
				Schema: map[string]*Schema{
					"name": &Schema{
						Type:     TypeString,
						Required: true,
					},
					"size": &Schema{
						Type:     TypeInt,
						Optional: true,
					},
					"green": &Schema{
						Type:     TypeBool,
						Optional: true,
						Computed: true,
					},
					"upside_down": &Schema{
						Type:     TypeBool,
						Computed: true,
					},
				},
			},
			Value: map[string]any{
				"name":  "my-fun-database",
				"size":  12,
				"green": true,
			},
			Expected: "green:1;name:my-fun-database;size:12;",
		},

		// test TypeMap nested in Schema: GH-7091
		testCase{
			Schema: &Resource{
				Schema: map[string]*Schema{
					"outer": &Schema{
						Type:     TypeSet,
						Required: true,
						Elem: &Schema{
							Type:     TypeMap,
							Optional: true,
						},
					},
				},
			},
			Value: map[string]any{
				"outer": NewSet(func(i any) int { return 42 }, []any{
					map[string]any{
						"foo": "bar",
						"baz": "foo",
					},
				}),
			},
			Expected: "outer:{[baz:foo;foo:bar;];};",
		},
	}

	for _, test := range tests {
		var gotBuf bytes.Buffer
		schema := test.Schema

		switch s := schema.(type) {
		case *Schema:
			SerializeValueForHash(&gotBuf, test.Value, s)
		case *Resource:
			SerializeResourceForHash(&gotBuf, test.Value, s)
		}

		got := gotBuf.String()
		if got != test.Expected {
			t.Errorf("hash(%#v) got %#v, but want %#v", test.Value, got, test.Expected)
		}
	}
}

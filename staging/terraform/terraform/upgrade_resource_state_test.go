// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package terraform

import (
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestStripRemovedStateAttributes(t *testing.T) {
	cases := []struct {
		name     string
		state    map[string]any
		expect   map[string]any
		ty       cty.Type
		modified bool
	}{
		{
			"removed string",
			map[string]any{
				"a": "ok",
				"b": "gone",
			},
			map[string]any{
				"a": "ok",
			},
			cty.Object(map[string]cty.Type{
				"a": cty.String,
			}),
			true,
		},
		{
			"removed null",
			map[string]any{
				"a": "ok",
				"b": nil,
			},
			map[string]any{
				"a": "ok",
			},
			cty.Object(map[string]cty.Type{
				"a": cty.String,
			}),
			true,
		},
		{
			"removed nested string",
			map[string]any{
				"a": "ok",
				"b": map[string]any{
					"a": "ok",
					"b": "removed",
				},
			},
			map[string]any{
				"a": "ok",
				"b": map[string]any{
					"a": "ok",
				},
			},
			cty.Object(map[string]cty.Type{
				"a": cty.String,
				"b": cty.Object(map[string]cty.Type{
					"a": cty.String,
				}),
			}),
			true,
		},
		{
			"removed nested list",
			map[string]any{
				"a": "ok",
				"b": map[string]any{
					"a": "ok",
					"b": []any{"removed"},
				},
			},
			map[string]any{
				"a": "ok",
				"b": map[string]any{
					"a": "ok",
				},
			},
			cty.Object(map[string]cty.Type{
				"a": cty.String,
				"b": cty.Object(map[string]cty.Type{
					"a": cty.String,
				}),
			}),
			true,
		},
		{
			"removed keys in set of objs",
			map[string]any{
				"a": "ok",
				"b": map[string]any{
					"a": "ok",
					"set": []any{
						map[string]any{
							"x": "ok",
							"y": "removed",
						},
						map[string]any{
							"x": "ok",
							"y": "removed",
						},
					},
				},
			},
			map[string]any{
				"a": "ok",
				"b": map[string]any{
					"a": "ok",
					"set": []any{
						map[string]any{
							"x": "ok",
						},
						map[string]any{
							"x": "ok",
						},
					},
				},
			},
			cty.Object(map[string]cty.Type{
				"a": cty.String,
				"b": cty.Object(map[string]cty.Type{
					"a": cty.String,
					"set": cty.Set(cty.Object(map[string]cty.Type{
						"x": cty.String,
					})),
				}),
			}),
			true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			modified := removeRemovedAttrs(tc.state, tc.ty)
			if !reflect.DeepEqual(tc.state, tc.expect) {
				t.Fatalf("expected: %#v\n      got: %#v\n", tc.expect, tc.state)
			}
			if modified != tc.modified {
				t.Fatal("incorrect return value")
			}
		})
	}
}

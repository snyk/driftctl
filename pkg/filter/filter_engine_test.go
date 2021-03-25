package filter

import (
	"errors"
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"

	testresource "github.com/cloudskiff/driftctl/test/resource"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func TestFilterEngine_Run(t *testing.T) {
	tests := []struct {
		name       string
		expr       string
		resources  []resource.Resource
		want       []resource.Resource
		compileErr error
		err        error
	}{
		{
			name:       "invalid expression (fail during expr compilation)",
			expr:       "???",
			compileErr: errors.New("SyntaxError: Unknown char: '?'"),
		},
		{
			name: "filter on type",
			expr: "Type=='filtered_resource'",
			resources: []resource.Resource{
				&testresource.FakeResource{},
				&testresource.FakeResource{
					Type: "filtered_resource",
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Type: "filtered_resource",
				},
			},
		},
		{
			name: "exclude all resource of type",
			expr: "Type!='filtered_resource'",
			resources: []resource.Resource{
				&testresource.FakeResource{},
				&testresource.FakeResource{
					Type: "filtered_resource",
				},
				&testresource.FakeResource{
					Type: "filtered_resource",
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{},
			},
		},
		{
			name: "filter on id",
			expr: "Id=='foobar'",
			resources: []resource.Resource{
				&testresource.FakeResource{
					Id: "barfoo",
				},
				&testresource.FakeResource{
					Id: "foobar",
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Id: "foobar",
				},
			},
		},
		{
			name: "filter on id and type",
			expr: "Id=='foobar' && Type=='filtered_resource'",
			resources: []resource.Resource{
				&testresource.FakeResource{
					Id: "foobar",
				},
				&testresource.FakeResource{
					Id:   "foobar",
					Type: "filtered_resource",
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Id:   "foobar",
					Type: "filtered_resource",
				},
			},
		},
		{
			name: "filter on resource field",
			expr: "Attr.bar_foo=='filtered'",
			resources: []resource.Resource{
				&testresource.FakeResource{},
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"bar_foo": cty.StringVal("filtered"),
						})
						return &v
					}(),
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"bar_foo": cty.StringVal("filtered"),
						})
						return &v
					}(),
				},
			},
		},
		{
			name: "filter on resource type and field",
			expr: "Type=='filtered_resource' && Attr.bar_foo=='filtered'",
			resources: []resource.Resource{
				&testresource.FakeResource{},
				&testresource.FakeResource{
					Type: "filtered_resource",
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"bar_foo": cty.StringVal("filtered"),
						})
						return &v
					}(),
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Type: "filtered_resource",
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"bar_foo": cty.StringVal("filtered"),
						})
						return &v
					}(),
				},
			},
		},
		{
			name: "filter on resource map of native type field",
			expr: "Attr.tags.foo=='foo'",
			resources: []resource.Resource{
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"tags": cty.MapVal(map[string]cty.Value{
								"foo": cty.StringVal("bar"),
								"bar": cty.StringVal("foo"),
							}),
						})
						return &v
					}(),
				},
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"tags": cty.MapVal(map[string]cty.Value{
								"foo": cty.StringVal("foo"),
								"bar": cty.StringVal("bar"),
							}),
						})
						return &v
					}(),
				},
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"tags": cty.MapVal(map[string]cty.Value{
								"foo": cty.StringVal("foo"),
								"bar": cty.StringVal("foo"),
							}),
						})
						return &v
					}(),
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"tags": cty.MapVal(map[string]cty.Value{
								"foo": cty.StringVal("foo"),
								"bar": cty.StringVal("bar"),
							}),
						})
						return &v
					}(),
				},
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"tags": cty.MapVal(map[string]cty.Value{
								"foo": cty.StringVal("foo"),
								"bar": cty.StringVal("foo"),
							}),
						})
						return &v
					}(),
				},
			},
		},
		{
			name: "filter on resource map of custom type field",
			expr: "Attr.custom_map.test.tag=='foo'",
			resources: []resource.Resource{
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"custom_map": cty.MapVal(map[string]cty.Value{
								"test": cty.ObjectVal(map[string]cty.Value{
									"tag": cty.StringVal("foo"),
								}),
							}),
						})
						return &v
					}(),
				},
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"custom_map": cty.MapVal(map[string]cty.Value{
								"test": cty.ObjectVal(map[string]cty.Value{
									"tag": cty.StringVal("bar"),
								}),
							}),
						})
						return &v
					}(),
				},
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"custom_map": cty.MapVal(map[string]cty.Value{
								"bar": cty.ObjectVal(map[string]cty.Value{
									"tag": cty.StringVal("foo"),
								}),
							}),
						})
						return &v
					}(),
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"custom_map": cty.MapVal(map[string]cty.Value{
								"test": cty.ObjectVal(map[string]cty.Value{
									"tag": cty.StringVal("foo"),
								}),
							}),
						})
						return &v
					}(),
				},
			},
		},
		{
			name: "filter on resource field array contains",
			expr: "Attr.slice[?contains(@, 'd')]",
			resources: []resource.Resource{
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"slice": cty.SetVal([]cty.Value{
								cty.StringVal("a"),
								cty.StringVal("b"),
								cty.StringVal("c"),
							}),
						})
						return &v
					}(),
				},
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"slice": cty.SetVal([]cty.Value{
								cty.StringVal("a"),
								cty.StringVal("b"),
								cty.StringVal("c"),
								cty.StringVal("d"),
							}),
						})
						return &v
					}(),
				},
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"slice": cty.SetVal([]cty.Value{
								cty.StringVal("d"),
							}),
						})
						return &v
					}(),
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"slice": cty.SetVal([]cty.Value{
								cty.StringVal("a"),
								cty.StringVal("b"),
								cty.StringVal("c"),
								cty.StringVal("d"),
							}),
						})
						return &v
					}(),
				},
				&testresource.FakeResource{
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"slice": cty.SetVal([]cty.Value{
								cty.StringVal("d"),
							}),
						})
						return &v
					}(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := BuildExpression(tt.expr)
			if tt.compileErr != nil && err == nil {
				t.Fatal("Expected compileErr got nil")
			}
			if err != nil && err.Error() != tt.compileErr.Error() {
				t.Fatalf("BuildExpression() error = '%s', want '%s'", err, tt.compileErr)
			}
			e := NewFilterEngine(expr)
			got, err := e.Run(tt.resources)
			if tt.err != nil && err == nil {
				t.Fatal("Expected err got nil")
			}
			if tt.err != nil && err.Error() != tt.err.Error() {
				t.Errorf("Run() error = %s, wantErr %s", err, tt.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Run() got = %v, want %v", got, tt.want)
			}
		})
	}
}

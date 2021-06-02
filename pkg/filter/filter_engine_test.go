package filter

import (
	"errors"
	"reflect"
	"testing"

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
				&testresource.FakeResource{
					Attrs: &resource.Attributes{},
				},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{},
					Type:  "filtered_resource",
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{},
					Type:  "filtered_resource",
				},
			},
		},
		{
			name: "exclude all resource of type",
			expr: "Type!='filtered_resource'",
			resources: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{}},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{},
					Type:  "filtered_resource",
				},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{},
					Type:  "filtered_resource",
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{}},
			},
		},
		{
			name: "filter on id",
			expr: "Id=='foobar'",
			resources: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{},
					Id:    "barfoo",
				},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{},
					Id:    "foobar",
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{},
					Id:    "foobar",
				},
			},
		},
		{
			name: "filter on id and type",
			expr: "Id=='foobar' && Type=='filtered_resource'",
			resources: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{},
					Id:    "foobar",
				},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{},
					Id:    "foobar",
					Type:  "filtered_resource",
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{},
					Id:    "foobar",
					Type:  "filtered_resource",
				},
			},
		},
		{
			name: "filter on resource field",
			expr: "Attr.bar_foo=='filtered'",
			resources: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{}},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"bar_foo": "filtered",
					},
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"bar_foo": "filtered",
					},
				},
			},
		},
		{
			name: "filter on resource type and field",
			expr: "Type=='filtered_resource' && Attr.bar_foo=='filtered'",
			resources: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{},
				},
				&testresource.FakeResource{
					Type: "filtered_resource",
					Attrs: &resource.Attributes{
						"bar_foo": "filtered",
					},
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Type: "filtered_resource",
					Attrs: &resource.Attributes{
						"bar_foo": "filtered",
					},
				},
			},
		},
		{
			name: "filter on resource map of native type field",
			expr: "Attr.tags.foo=='foo'",
			resources: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"tags": map[string]interface{}{
							"foo": "bar",
							"bar": "foo",
						},
					},
				},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"tags": map[string]interface{}{
							"foo": "foo",
							"bar": "bar",
						},
					},
				},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"tags": map[string]interface{}{
							"foo": "foo",
							"bar": "foo",
						},
					},
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"tags": map[string]interface{}{
							"foo": "foo",
							"bar": "bar",
						},
					},
				},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"tags": map[string]interface{}{
							"foo": "foo",
							"bar": "foo",
						},
					},
				},
			},
		},
		{
			name: "filter on resource map of custom type field",
			expr: "Attr.custom_map.test.tag=='foo'",
			resources: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"custom_map": map[string]interface{}{
							"test": map[string]interface{}{
								"tag": "foo",
							},
						},
					},
				},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"custom_map": map[string]interface{}{
							"test": map[string]interface{}{
								"tag": "bar",
							},
						},
					},
				},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"custom_map": map[string]interface{}{
							"bar": map[string]interface{}{
								"tag": "foo",
							},
						},
					},
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"custom_map": map[string]interface{}{
							"test": map[string]interface{}{
								"tag": "foo",
							},
						},
					},
				},
			},
		},
		{
			name: "filter on resource field array contains",
			expr: "Attr.slice[?contains(@, 'd')]",
			resources: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"slice": []string{"a", "b", "c"},
					},
				},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"slice": []string{"a", "b", "c", "d"},
					},
				},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"slice": []string{"d"},
					},
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"slice": []string{"a", "b", "c", "d"},
					},
				},
				&testresource.FakeResource{
					Attrs: &resource.Attributes{
						"slice": []string{"d"},
					},
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

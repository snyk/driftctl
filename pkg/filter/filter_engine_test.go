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
			expr: "Attr.BarFoo=='filtered'",
			resources: []resource.Resource{
				&testresource.FakeResource{},
				&testresource.FakeResource{
					BarFoo: "filtered",
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					BarFoo: "filtered",
				},
			},
		},
		{
			name: "filter on resource type and field",
			expr: "Type=='filtered_resource' && Attr.BarFoo=='filtered'",
			resources: []resource.Resource{
				&testresource.FakeResource{},
				&testresource.FakeResource{
					Type:   "filtered_resource",
					BarFoo: "filtered",
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Type:   "filtered_resource",
					BarFoo: "filtered",
				},
			},
		},
		{
			name: "filter on resource map of native type field",
			expr: "Attr.Tags.foo=='foo'",
			resources: []resource.Resource{
				&testresource.FakeResource{
					Tags: map[string]string{
						"foo": "bar",
						"bar": "foo",
					},
				},
				&testresource.FakeResource{
					Tags: map[string]string{
						"foo": "foo",
						"bar": "bar",
					},
				},
				&testresource.FakeResource{
					Tags: map[string]string{
						"foo": "foo",
						"bar": "foo",
					},
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Tags: map[string]string{
						"foo": "foo",
						"bar": "bar",
					},
				},
				&testresource.FakeResource{
					Tags: map[string]string{
						"foo": "foo",
						"bar": "foo",
					},
				},
			},
		},
		{
			name: "filter on resource map of custom type field",
			expr: "Attr.CustomMap.test.Tag=='foo'",
			resources: []resource.Resource{
				&testresource.FakeResource{
					CustomMap: map[string]struct{ Tag string }{
						"test": {
							Tag: "foo",
						},
					},
				},
				&testresource.FakeResource{
					CustomMap: map[string]struct{ Tag string }{
						"test": {
							Tag: "bar",
						},
					},
				},
				&testresource.FakeResource{
					CustomMap: map[string]struct{ Tag string }{
						"bar": {
							Tag: "foo",
						},
					},
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					CustomMap: map[string]struct{ Tag string }{
						"test": {
							Tag: "foo",
						},
					},
				},
			},
		},
		{
			name: "filter on resource field array contains",
			expr: "Attr.Slice[?contains(@, 'd')]",
			resources: []resource.Resource{
				&testresource.FakeResource{
					Slice: []string{"a", "b", "c"},
				},
				&testresource.FakeResource{
					Slice: []string{"a", "b", "c", "d"},
				},
				&testresource.FakeResource{
					Slice: []string{"d"},
				},
			},
			want: []resource.Resource{
				&testresource.FakeResource{
					Slice: []string{"a", "b", "c", "d"},
				},
				&testresource.FakeResource{
					Slice: []string{"d"},
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

package filter

import (
	"os"
	"path"
	"reflect"
	"testing"

	resource2 "github.com/cloudskiff/driftctl/test/resource"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func TestDriftIgnore_Run(t *testing.T) {
	tests := []struct {
		name      string
		resources []resource.Resource
		want      []resource.Resource
	}{
		{
			name: "drift_ignore_no_file",
			resources: []resource.Resource{
				resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
			},
			want: []resource.Resource{
				resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
			},
		},
		{
			name: "drift_ignore_empty",
			resources: []resource.Resource{
				resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
			},
			want: []resource.Resource{
				resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
			},
		},
		{
			name: "drift_ignore_invalid_lines",
			resources: []resource.Resource{
				resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
				resource2.FakeResource{
					Type: "ignored_resource",
					Id:   "id2",
				},
			},
			want: []resource.Resource{
				resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
			},
		},
		{
			name: "drift_ignore_valid",
			resources: []resource.Resource{
				resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
				resource2.FakeResource{
					Type: "wildcard_resource",
					Id:   "id1",
				},
				resource2.FakeResource{
					Type: "wildcard_resource",
					Id:   "id2",
				},
				resource2.FakeResource{
					Type: "wildcard_resource",
					Id:   "id3",
				},
				resource2.FakeResource{
					Type: "ignored_resource",
					Id:   "id2",
				},
				resource2.FakeResource{
					Type: "resource_type",
					Id:   "id.with.dots",
				},
			},
			want: []resource.Resource{
				resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cwd, _ := os.Getwd()
			defer func() { _ = os.Chdir(cwd) }()
			if err := os.Chdir(path.Join("testdata", tt.name)); err != nil {
				t.Fatal(err)
			}
			r := NewDriftIgnore()
			if got := r.Run(tt.resources); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Run() = %v, want %v", got, tt.want)
			}
		})
	}
}

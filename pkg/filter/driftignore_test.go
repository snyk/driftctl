package filter

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/r3labs/diff/v2"

	resource2 "github.com/cloudskiff/driftctl/test/resource"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func TestDriftIgnore_FilterResources(t *testing.T) {
	tests := []struct {
		name      string
		resources []resource.Resource
		want      []resource.Resource
	}{
		{
			name: "drift_ignore_no_file",
			resources: []resource.Resource{
				&resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
			},
			want: []resource.Resource{
				&resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
			},
		},
		{
			name: "drift_ignore_empty",
			resources: []resource.Resource{
				&resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
			},
			want: []resource.Resource{
				&resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
			},
		},
		{
			name: "drift_ignore_invalid_lines",
			resources: []resource.Resource{
				&resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
				&resource2.FakeResource{
					Type: "ignored_resource",
					Id:   "id2",
				},
			},
			want: []resource.Resource{
				&resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
			},
		},
		{
			name: "drift_ignore_valid",
			resources: []resource.Resource{
				&resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
				&resource2.FakeResource{
					Type: "wildcard_resource",
					Id:   "id1",
				},
				&resource2.FakeResource{
					Type: "wildcard_resource",
					Id:   "id2",
				},
				&resource2.FakeResource{
					Type: "wildcard_resource",
					Id:   "id3",
				},
				&resource2.FakeResource{
					Type: "ignored_resource",
					Id:   "id2",
				},
				&resource2.FakeResource{
					Type: "resource_type",
					Id:   "id.with.dots",
				},
				resource2.FakeResource{
					Type: "resource_type",
					Id:   "idwith\\",
				},
				resource2.FakeResource{
					Type: "resource_type",
					Id:   "idwith\\backslashes",
				},
			},
			want: []resource.Resource{
				&resource2.FakeResource{
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
			got := r.FilterResources(tt.resources)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDriftIgnore_FilterDrift(t *testing.T) {

	tests := []struct {
		name       string
		difference []analyser.Difference
		want       []analyser.Difference
	}{
		{
			name: "drift_ignore_no_file",
			difference: []analyser.Difference{
				{
					Res: resource2.FakeResource{Type: "type1", Id: "id1"},
					Changelog: []diff.Change{
						{Path: []string{"Id"}},
					},
				},
				{
					Res: resource2.FakeResource{Type: "type2", Id: "id2"},
					Changelog: []diff.Change{
						{Path: []string{"Id"}},
					},
				},
			},
			want: []analyser.Difference{
				{
					Res: resource2.FakeResource{Type: "type1", Id: "id1"},
					Changelog: []diff.Change{
						{Path: []string{"Id"}},
					},
				},
				{
					Res: resource2.FakeResource{Type: "type2", Id: "id2"},
					Changelog: []diff.Change{
						{Path: []string{"Id"}},
					},
				},
			},
		},
		{
			name: "drift_ignore_empty",
			difference: []analyser.Difference{
				{
					Res: resource2.FakeResource{Type: "type1", Id: "id1"},
					Changelog: []diff.Change{
						{Path: []string{"Id"}},
					},
				},
				{
					Res: resource2.FakeResource{Type: "type2", Id: "id2"},
					Changelog: []diff.Change{
						{Path: []string{"Id"}},
					},
				},
			},
			want: []analyser.Difference{
				{
					Res: resource2.FakeResource{Type: "type1", Id: "id1"},
					Changelog: []diff.Change{
						{Path: []string{"Id"}},
					},
				},
				{
					Res: resource2.FakeResource{Type: "type2", Id: "id2"},
					Changelog: []diff.Change{
						{Path: []string{"Id"}},
					},
				},
			},
		},
		{
			name: "drift_ignore_fields",
			difference: []analyser.Difference{
				{
					Res: resource2.FakeResource{Type: "res_type", Id: "full_drift_ignored"},
					Changelog: []diff.Change{
						{Path: []string{"Json"}},
						{Path: []string{"Foobar"}},
					},
				},
				{
					Res: resource2.FakeResource{Type: "res_type", Id: "partial_drift_ignored"},
					Changelog: []diff.Change{
						{Path: []string{"Json"}},
						{Path: []string{"Foobar"}},
					},
				},
				{
					Res: resource2.FakeResource{Type: "resource_type", Id: "id.with.dots"},
					Changelog: []diff.Change{
						{Path: []string{"Json"}},
						{Path: []string{"Foobar"}},
					},
				},
				{
					Res: resource2.FakeResource{Type: "resource_type", Id: "idwith\\"},
					Changelog: []diff.Change{
						{Path: []string{"Json"}},
					},
				},
				{
					Res: resource2.FakeResource{Type: "resource_type", Id: "idwith\\backslashes"},
					Changelog: []diff.Change{
						{Path: []string{"Json"}},
						{Path: []string{"Foobar"}},
					},
				},
				{
					Res: resource2.FakeResource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Changelog: []diff.Change{
						{Path: []string{"Struct", "Baz"}},
						{Path: []string{"Struct", "Bar"}},
					},
				},
				{
					Res: resource2.FakeResource{Type: "res_type", Id: "endofpath_drift_ignored"},
					Changelog: []diff.Change{
						{Path: []string{"Struct", "Baz"}},
						{Path: []string{"Struct", "Bar"}},
					},
				},
			},
			want: []analyser.Difference{
				{
					Res: resource2.FakeResource{Type: "res_type", Id: "partial_drift_ignored"},
					Changelog: []diff.Change{
						{Path: []string{"Json"}},
					},
				},
				{
					Res: resource2.FakeResource{Type: "resource_type", Id: "id.with.dots"},
					Changelog: []diff.Change{
						{Path: []string{"Foobar"}},
					},
				},
				{
					Res: resource2.FakeResource{Type: "resource_type", Id: "idwith\\backslashes"},
					Changelog: []diff.Change{
						{Path: []string{"Json"}},
					},
				},
				{
					Res: resource2.FakeResource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Changelog: []diff.Change{
						{Path: []string{"Struct", "Bar"}},
					},
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
			got := r.FilterDrift(tt.difference)
			assert.Equal(t, tt.want, got)

		})
	}
}

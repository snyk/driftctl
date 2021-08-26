package filter

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func TestDriftIgnore_IsResourceIgnored(t *testing.T) {
	tests := []struct {
		name      string
		resources []*resource.Resource
		want      []bool
		path      string
	}{
		{
			name: "drift_ignore_no_file",
			resources: []*resource.Resource{
				{
					Type: "type1",
					Id:   "id1",
				},
			},
			want: []bool{
				false,
			},
			path: "testdata/drift_ignore_no_file/.driftignore",
		},
		{
			name: "drift_ignore_empty",
			resources: []*resource.Resource{
				{
					Type: "type1",
					Id:   "id1",
				},
			},
			want: []bool{
				false,
			},
			path: "testdata/drift_ignore_empty/.driftignore",
		},
		{
			name: "drift_ignore_invalid_lines",
			resources: []*resource.Resource{
				{
					Type: "type1",
					Id:   "id1",
				},
				{
					Type: "ignored_resource",
					Id:   "id2",
				},
			},
			want: []bool{
				false,
				true,
			},
			path: "testdata/drift_ignore_invalid_lines/.driftignore",
		},
		{
			name: "drift_ignore_valid",
			resources: []*resource.Resource{
				{
					Type: "type1",
					Id:   "id1",
				},
				{
					Type: "wildcard_resource",
					Id:   "id1/with/slash",
				},
				{
					Type: "wildcard_resource",
					Id:   "id1",
				},
				{
					Type: "wildcard_resource",
					Id:   "id2",
				},
				{
					Type: "wildcard_resource",
					Id:   "id3",
				},
				{
					Type: "ignored_resource",
					Id:   "id2",
				},
				{
					Type: "resource_type",
					Id:   "id.with.dots",
				},
				{
					Type: "resource_type",
					Id:   "idwith\\",
				},
				{
					Type: "resource_type",
					Id:   "idwith\\backslashes",
				},
				{
					Type: "resource_type",
					Id:   "idwith/slashes",
				},
			},
			want: []bool{
				false,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
			},
			path: "testdata/drift_ignore_valid/.driftignore",
		},
		{
			name: "drift_ignore_wildcard",
			resources: []*resource.Resource{
				{
					Type: "type1",
					Id:   "id1",
				},
				{
					Type: "type2",
					Id:   "id1",
				},
				{
					Type: "type2",
					Id:   "id11",
				},
				{
					Type: "type2",
					Id:   "id2",
				},
				{
					Type: "type3",
					Id:   "id100",
				},
				{
					Type: "type3",
					Id:   "id101",
				},
				{
					Type: "type4",
					Id:   "id\\WithBac*slash***\\*\\",
				},
			},
			want: []bool{
				false,
				true,
				true,
				false,
				true,
				false,
				true,
			},
			path: "testdata/drift_ignore_wildcard/.driftignore",
		},
		{
			name: "drift_ignore_all_exclude",
			resources: []*resource.Resource{
				{
					Type: "type1",
					Id:   "id1",
				},
				{
					Type: "type2",
					Id:   "id1",
				},
				{
					Type: "type2",
					Id:   "id11",
				},
				{
					Type: "type2",
					Id:   "id2",
				},
				{
					Type: "type3",
					Id:   "id100",
				},
				{
					Type: "type3",
					Id:   "id101",
				},
				{
					Type: "iam_user",
					Id:   "id\\WithBac*slash***\\*\\",
				},
				{
					Type: "some_type",
					Id:   "idwith/slash",
				},
				{
					Type: "some_type",
					Id:   "idwith/slash/",
				},
			},
			want: []bool{
				true,
				true,
				true,
				true,
				true,
				true,
				false,
				false,
				true,
			},
			path: "testdata/drift_ignore_all_exclude/.driftignore",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cwd, _ := os.Getwd()
			defer func() { _ = os.Chdir(cwd) }()

			r := NewDriftIgnore(tt.path)
			got := make([]bool, 0, len(tt.want))
			for _, res := range tt.resources {
				got = append(got, r.IsResourceIgnored(res))
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDriftIgnore_IsFieldIgnored(t *testing.T) {

	type Args struct {
		Res  *resource.Resource
		Path []string
		Want bool
	}

	tests := []struct {
		name string
		args []Args
		path string
	}{
		{
			name: "drift_ignore_no_file",
			args: []Args{

				{
					Res:  &resource.Resource{Type: "type1", Id: "id1"},
					Path: []string{"Id"},
					Want: false,
				},
				{
					Res:  &resource.Resource{Type: "type2", Id: "id2"},
					Path: []string{"Id"},
					Want: false,
				},
			},
			path: "testdata/drift_ignore_no_file/.driftignore",
		},
		{
			name: "drift_ignore_empty",
			args: []Args{
				{
					Res:  &resource.Resource{Type: "type1", Id: "id1"},
					Path: []string{"Id"},
					Want: false,
				},
				{
					Res:  &resource.Resource{Type: "type2", Id: "id2"},
					Path: []string{"Id"},
					Want: false,
				},
			},
			path: "testdata/drift_ignore_empty/.driftignore",
		},
		{
			name: "drift_ignore_fields",
			args: []Args{
				{
					Res:  &resource.Resource{Type: "res_type", Id: "full_drift_ignored"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "full_drift_ignored"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "partial_drift_ignored"},
					Path: []string{"json"},
					Want: false,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "partial_drift_ignored"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "id.with.dots"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "id.with.dots"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\backslashes"},
					Path: []string{"json"},
					Want: false,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\backslashes"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Path: []string{"struct", "baz"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Path: []string{"struct", "bar"},
					Want: false,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "endofpath_drift_ignored"},
					Path: []string{"struct", "baz"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "endofpath_drift_ignored"},
					Path: []string{"struct", "bar"},
					Want: true,
				},
			},
			path: "testdata/drift_ignore_fields/.driftignore",
		},
		{
			name: "drift_ignore_all_exclude_field",
			args: []Args{
				{
					Res:  &resource.Resource{Type: "res_type", Id: "full_drift_ignored"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "full_drift_ignored"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "partial_drift_ignored"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "partial_drift_ignored"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "id.with.dots"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "id.with.dots"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\backslashes"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\backslashes"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Path: []string{"struct", "baz"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Path: []string{"struct", "bar"},
					Want: false,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "endofpath_drift_ignored"},
					Path: []string{"struct", "baz"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "endofpath_drift_ignored"},
					Path: []string{"struct", "bar"},
					Want: false,
				},
			},
			path: "testdata/drift_ignore_all_exclude_field/.driftignore",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cwd, _ := os.Getwd()
			defer func() { _ = os.Chdir(cwd) }()

			r := NewDriftIgnore(tt.path)
			for _, arg := range tt.args {
				got := r.IsFieldIgnored(arg.Res, arg.Path)
				if arg.Want != got {
					t.Errorf("%s.%s.%s expected %v got %v", arg.Res.ResourceType(), arg.Res.ResourceId(), strings.Join(arg.Path, "."), arg.Want, got)
				}
			}
		})
	}
}

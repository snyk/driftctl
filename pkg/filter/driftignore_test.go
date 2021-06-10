package filter

import (
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	resource2 "github.com/cloudskiff/driftctl/test/resource"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func TestDriftIgnore_IsResourceIgnored(t *testing.T) {
	tests := []struct {
		name      string
		resources []resource.Resource
		want      []bool
	}{
		{
			name: "drift_ignore_no_file",
			resources: []resource.Resource{
				&resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
			},

			want: []bool{
				false,
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
			want: []bool{
				false,
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
			want: []bool{
				false,
				true,
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
				&resource2.FakeResource{
					Type: "resource_type",
					Id:   "idwith\\",
				},
				&resource2.FakeResource{
					Type: "resource_type",
					Id:   "idwith\\backslashes",
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
			},
		},
		{
			name: "drift_ignore_wildcard",
			resources: []resource.Resource{
				&resource2.FakeResource{
					Type: "type1",
					Id:   "id1",
				},
				&resource2.FakeResource{
					Type: "type2",
					Id:   "id1",
				},
				&resource2.FakeResource{
					Type: "type2",
					Id:   "id11",
				},
				&resource2.FakeResource{
					Type: "type2",
					Id:   "id2",
				},
				&resource2.FakeResource{
					Type: "type3",
					Id:   "id100",
				},
				&resource2.FakeResource{
					Type: "type3",
					Id:   "id101",
				},
				&resource2.FakeResource{
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
		Res  resource.Resource
		Path []string
		Want bool
	}

	tests := []struct {
		name string
		args []Args
	}{
		{
			name: "drift_ignore_no_file",
			args: []Args{

				{
					Res:  &resource2.FakeResource{Type: "type1", Id: "id1"},
					Path: []string{"Id"},
					Want: false,
				},
				{
					Res:  &resource2.FakeResource{Type: "type2", Id: "id2"},
					Path: []string{"Id"},
					Want: false,
				},
			},
		},
		{
			name: "drift_ignore_empty",
			args: []Args{
				{
					Res:  &resource2.FakeResource{Type: "type1", Id: "id1"},
					Path: []string{"Id"},
					Want: false,
				},
				{
					Res:  &resource2.FakeResource{Type: "type2", Id: "id2"},
					Path: []string{"Id"},
					Want: false,
				},
			},
		},
		{
			name: "drift_ignore_fields",
			args: []Args{
				{
					Res:  &resource2.FakeResource{Type: "res_type", Id: "full_drift_ignored"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource2.FakeResource{Type: "res_type", Id: "full_drift_ignored"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource2.FakeResource{Type: "res_type", Id: "partial_drift_ignored"},
					Path: []string{"json"},
					Want: false,
				},
				{
					Res:  &resource2.FakeResource{Type: "res_type", Id: "partial_drift_ignored"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource2.FakeResource{Type: "resource_type", Id: "id.with.dots"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource2.FakeResource{Type: "resource_type", Id: "id.with.dots"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource2.FakeResource{Type: "resource_type", Id: "idwith\\"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource2.FakeResource{Type: "resource_type", Id: "idwith\\backslashes"},
					Path: []string{"json"},
					Want: false,
				},
				{
					Res:  &resource2.FakeResource{Type: "resource_type", Id: "idwith\\backslashes"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource2.FakeResource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Path: []string{"struct", "baz"},
					Want: true,
				},
				{
					Res:  &resource2.FakeResource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Path: []string{"struct", "bar"},
					Want: false,
				},
				{
					Res:  &resource2.FakeResource{Type: "res_type", Id: "endofpath_drift_ignored"},
					Path: []string{"struct", "baz"},
					Want: true,
				},
				{
					Res:  &resource2.FakeResource{Type: "res_type", Id: "endofpath_drift_ignored"},
					Path: []string{"struct", "bar"},
					Want: true,
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
			for _, arg := range tt.args {
				got := r.IsFieldIgnored(arg.Res, arg.Path)
				if arg.Want != got {
					t.Errorf("%s.%s.%s expected %v got %v", arg.Res.TerraformType(), arg.Res.TerraformId(), strings.Join(arg.Path, "."), arg.Want, got)
				}
			}
		})
	}
}

func Test_escapableSplit(t *testing.T) {
	tests := []struct {
		name string
		line string
		want []string
	}{
		{
			name: "Dot at start",
			line: ".",
			want: []string{"."},
		},
		{
			name: "Dot at end",
			line: "test.toto.",
			want: []string{"test", "toto"},
		},
		{
			name: "wildcard dot",
			line: "*.subfoobar",
			want: []string{"*", "subfoobar"},
		},
		{
			name: "text wildcard dot",
			line: "res*.subfoobar",
			want: []string{"res*", "subfoobar"},
		},
		{
			name: "missing text multiple wildcard dot",
			line: "r*s*.s**ub***ob********a*r",
			want: []string{"r*s*", "s*ub*ob*a*r"},
		},
		{
			name: "prefix wildcard dot",
			line: "*res.subfoobar",
			want: []string{"*res", "subfoobar"},
		},
		{
			name: "suffix multiple wildcard dot",
			line: "res.subfoobar*****",
			want: []string{"res", "subfoobar*"},
		},
		{
			name: "dot wildcard",
			line: "res.*",
			want: []string{"res", "*"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readDriftIgnoreLine(tt.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EscapableSplit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDriftIgnore_IsTypeIgnored(t *testing.T) {
	tests := []struct {
		name  string
		types []string
		want  []bool
	}{
		{
			name: "drift_ignore_valid",
			types: []string{
				"wildcard_resource",
				"no_ignored",
			},
			want: []bool{
				true,
				false,
			},
		},
		{
			name: "drift_ignore_wildcard",
			types: []string{
				"wildcard_resource",
				"no_ignored",
			},
			want: []bool{
				false,
				false,
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
			got := make([]bool, 0, len(tt.want))
			for _, res := range tt.types {
				got = append(got, r.IsTypeIgnored(res))
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

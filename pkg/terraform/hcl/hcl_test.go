package hcl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHCL_getCurrentWorkspaceName(t *testing.T) {
	cases := []struct {
		name string
		dir  string
		want string
	}{
		{
			name: "test with non-default workspace",
			dir:  "testdata/foo_workspace",
			want: "foo",
		},
		{
			name: "test with non-existing directory",
			dir:  "testdata/noenvfile",
			want: "default",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			workspace := GetCurrentWorkspaceName(tt.dir)
			assert.Equal(t, tt.want, workspace)
		})
	}
}

package hcl

import (
	"testing"

	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/stretchr/testify/assert"
)

func TestCloud_SupplierConfig(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		want     *config.SupplierConfig
		wantErr  string
	}{
		{
			name:     "test with cloud block and default workspace",
			filename: "testdata/cloud_block.tf",
			want: &config.SupplierConfig{
				Key:     "tfstate",
				Backend: "tfcloud",
				Path:    "example_corp/default",
			},
		},
		{
			name:     "test with cloud block and non-default workspace",
			filename: "testdata/cloud_block_workspace.tf",
			want: &config.SupplierConfig{
				Key:     "tfstate",
				Backend: "tfcloud",
				Path:    "example_corp/my-workspace",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			hcl, err := ParseTerraformFromHCL(tt.filename)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			if hcl.Cloud == nil {
				assert.Nil(t, tt.want)
				return
			}

			assert.Equal(t, *tt.want, *hcl.Cloud.SupplierConfig("default"))
		})
	}
}

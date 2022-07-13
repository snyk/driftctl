package hcl

import (
	"path"
	"testing"

	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/stretchr/testify/assert"
)

func TestBackend_SupplierConfig(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		want     *config.SupplierConfig
		wantErr  string
	}{
		{
			name:     "test with no backend block",
			filename: "testdata/no_backend_block.tf",
			want:     nil,
		},
		{
			name:     "test with local backend block",
			filename: "testdata/local_backend_block.tf",
			want: &config.SupplierConfig{
				Key:  "tfstate",
				Path: "terraform-state-prod/network/terraform.tfstate",
			},
		},
		{
			name:     "test with S3 backend block",
			filename: "testdata/s3_backend_block.tf",
			want: &config.SupplierConfig{
				Key:     "tfstate",
				Backend: "s3",
				Path:    "terraform-state-prod/network/terraform.tfstate",
			},
		},
		{
			name:     "test with S3 backend block with non-default workspace",
			filename: "testdata/s3_backend_workspace/s3_backend_block.tf",
			want: &config.SupplierConfig{
				Key:     "tfstate",
				Backend: "s3",
				Path:    "terraform-state-prod/env:/bar/network/terraform.tfstate",
			},
		},
		{
			name:     "test with GCS backend block",
			filename: "testdata/gcs_backend_block.tf",
			want: &config.SupplierConfig{
				Key:     "tfstate",
				Backend: "gs",
				Path:    "tf-state-prod/terraform/state/default.tfstate",
			},
		},
		{
			name:     "test with Azure backend block",
			filename: "testdata/azurerm_backend_block.tf",
			want: &config.SupplierConfig{
				Key:     "tfstate",
				Backend: "azurerm",
				Path:    "states/prod.terraform.tfstate",
			},
		},
		{
			name:     "test with Azure backend block with non-default workspace",
			filename: "testdata/azurerm_backend_workspace/azurerm_backend_block.tf",
			want: &config.SupplierConfig{
				Key:     "tfstate",
				Backend: "azurerm",
				Path:    "states/prod.terraform.tfstateenv:bar",
			},
		},
		{
			name:     "test with unknown backend",
			filename: "testdata/unknown_backend_block.tf",
			want:     nil,
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

			ws := GetCurrentWorkspaceName(path.Dir(tt.filename))
			if hcl.Backend == nil || hcl.Backend.SupplierConfig(ws) == nil {
				assert.Nil(t, tt.want)
				return
			}

			assert.Equal(t, *tt.want, *hcl.Backend.SupplierConfig(ws))
		})
	}
}

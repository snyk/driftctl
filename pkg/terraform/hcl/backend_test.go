package hcl

import (
	"testing"

	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/stretchr/testify/assert"
)

func TestBackend_SupplierConfig(t *testing.T) {
	cases := []struct {
		name    string
		dir     string
		want    *config.SupplierConfig
		wantErr string
	}{
		{
			name:    "test with no backend block",
			dir:     "testdata/no_backend_block.tf",
			want:    nil,
			wantErr: "testdata/no_backend_block.tf:1,11-11: Missing backend block; A backend block is required.",
		},
		{
			name: "test with local backend block",
			dir:  "testdata/local_backend_block.tf",
			want: &config.SupplierConfig{
				Key:  "tfstate",
				Path: "terraform-state-prod/network/terraform.tfstate",
			},
		},
		{
			name: "test with S3 backend block",
			dir:  "testdata/s3_backend_block.tf",
			want: &config.SupplierConfig{
				Key:     "tfstate",
				Backend: "s3",
				Path:    "terraform-state-prod/network/terraform.tfstate",
			},
		},
		{
			name: "test with GCS backend block",
			dir:  "testdata/gcs_backend_block.tf",
			want: &config.SupplierConfig{
				Key:     "tfstate",
				Backend: "gs",
				Path:    "tf-state-prod/terraform/state.tfstate",
			},
		},
		{
			name: "test with Azure backend block",
			dir:  "testdata/azurerm_backend_block.tf",
			want: &config.SupplierConfig{
				Key:     "tfstate",
				Backend: "azurerm",
				Path:    "states/prod.terraform.tfstate",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			hcl, err := ParseTerraformFromHCL(tt.dir)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			if hcl.Backend.SupplierConfig() == nil {
				assert.Nil(t, tt.want)
				return
			}

			assert.Equal(t, *tt.want, *hcl.Backend.SupplierConfig())
		})
	}
}

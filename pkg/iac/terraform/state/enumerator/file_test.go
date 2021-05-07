package enumerator

import (
	"reflect"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/iac/config"
	"github.com/stretchr/testify/assert"
)

func TestFileEnumerator_Enumerate(t *testing.T) {
	tests := []struct {
		name   string
		config config.SupplierConfig
		want   []string
		err    string
	}{
		{
			name: "subfolder nesting",
			config: config.SupplierConfig{
				Path: "testdata/states",
			},
			want: []string{
				"testdata/states/symlink.tfstate",
				"testdata/states/terraform.tfstate",
				"testdata/states/lambda/lambda.tfstate",
				"testdata/states/s3/terraform.tfstate",
				"testdata/states/symlink-to-s3-folder/terraform.tfstate",
			},
		},
		{
			name: "subfolder nesting glob",
			config: config.SupplierConfig{
				Path: "testdata/states/**/*.tfstate",
			},
			want: []string{
				"testdata/states/symlink.tfstate",
				"testdata/states/terraform.tfstate",
				"testdata/states/lambda/lambda.tfstate",
				"testdata/states/s3/terraform.tfstate",
				"testdata/states/symlink-to-s3-folder/terraform.tfstate",
			},
		},
		{
			name: "subfolder nesting glob upper directory",
			config: config.SupplierConfig{
				Path: "testdata/states/s3/../**/*.tfstate",
			},
			want: []string{
				"testdata/states/symlink.tfstate",
				"testdata/states/terraform.tfstate",
				"testdata/states/lambda/lambda.tfstate",
				"testdata/states/s3/terraform.tfstate",
				"testdata/states/symlink-to-s3-folder/terraform.tfstate",
			},
		},
		{
			name: "symlinked folder",
			config: config.SupplierConfig{
				Path: "testdata/symlink",
			},
			want: []string{
				"testdata/states/symlink.tfstate",
				"testdata/states/terraform.tfstate",
				"testdata/states/lambda/lambda.tfstate",
				"testdata/states/s3/terraform.tfstate",
				"testdata/states/symlink-to-s3-folder/terraform.tfstate",
			},
		},
		{
			name: "single state file",
			config: config.SupplierConfig{
				Path: "testdata/states/terraform.tfstate",
			},
			want: []string{
				"testdata/states/terraform.tfstate",
			},
		},
		{
			name: "single symlink state file",
			config: config.SupplierConfig{
				Path: "testdata/states/symlink.tfstate",
			},
			want: []string{
				"testdata/states/terraform.tfstate",
			},
		},
		{
			name: "invalid folder",
			config: config.SupplierConfig{
				Path: "/tmp/dummy-folder/that/does/not/exist",
			},
			want: nil,
			err:  "lstat /tmp/dummy-folder/that/does/not/exist: no such file or directory",
		},
		{
			name: "invalid symlink",
			config: config.SupplierConfig{
				Path: "testdata/invalid_symlink/invalid",
			},
			want: nil,
			err:  "lstat testdata/invalid_symlink/test: no such file or directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewFileEnumerator(tt.config)
			got, err := s.Enumerate()
			if tt.err != "" {
				assert.EqualError(t, err, tt.err)
			} else {
				assert.NoError(t, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Enumerate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

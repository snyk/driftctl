package supplier

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/iac/config"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

func TestGetIACSupplier(t *testing.T) {
	type args struct {
		config []config.SupplierConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "test unknown supplier",
			args: args{
				config: []config.SupplierConfig{
					{
						Key: "foobar",
					},
				},
			},
			wantErr: fmt.Errorf("Unsupported supplier 'foobar'"),
		},
		{
			name: "test unknown supplier in multiples states",
			args: args{
				config: []config.SupplierConfig{
					{
						Key: "foobar",
					},
					{
						Key:     "tfstate",
						Backend: "",
						Path:    "terraform.tfstate",
					},
				},
			},
			wantErr: fmt.Errorf("Unsupported supplier 'foobar'"),
		},
		{
			name: "test unknown backend",
			args: args{
				config: []config.SupplierConfig{
					{Key: "tfstate", Backend: "foobar"},
				},
			},
			wantErr: fmt.Errorf("Unsupported backend 'foobar'"),
		},
		{
			name: "test valid tfstate://terraform.tfstate",
			args: args{
				config: []config.SupplierConfig{
					{Key: "tfstate", Backend: "", Path: "terraform.tfstate"},
				},
			},
			wantErr: nil,
		},
		{
			name: "test valid multiples states",
			args: args{
				config: []config.SupplierConfig{
					{Key: "tfstate", Backend: "", Path: "terraform.tfstate"},
					{Key: "tfstate", Backend: "s3", Path: "terraform.tfstate"},
					{Key: "tfstate", Backend: "", Path: "terraform2.tfstate"},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetIACSupplier(tt.args.config, terraform.NewProviderLibrary())
			if tt.wantErr != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("GetIACSupplier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetSupportedSchemes(t *testing.T) {

	want := []string{
		"tfstate://",
		"tfstate+s3://",
	}

	if got := GetSupportedSchemes(); !reflect.DeepEqual(got, want) {
		t.Errorf("GetSupportedSchemes() = %v, want %v", got, want)
	}
}

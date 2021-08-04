package config

import "testing"

func TestSupplierConfig_String(t *testing.T) {
	tests := []struct {
		name   string
		config SupplierConfig
		want   string
	}{
		{
			name:   "test with empty config",
			config: SupplierConfig{},
			want:   "",
		},
		{
			name: "test with empty path",
			config: SupplierConfig{
				Key:     "tfstate",
				Backend: "s3",
				Path:    "",
			},
			want: "tfstate+s3://",
		},
		{
			name: "test valid config",
			config: SupplierConfig{
				Key:     "tfstate",
				Backend: "s3",
				Path:    "my-bucket/terraform.tfstate",
			},
			want: "tfstate+s3://my-bucket/terraform.tfstate",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

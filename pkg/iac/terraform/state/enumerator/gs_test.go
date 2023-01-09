package enumerator

import (
	"testing"

	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/stretchr/testify/assert"
)

func TestGSEnumerator_NewGSEnumerator(t *testing.T) {
	tests := []struct {
		name   string
		config config.SupplierConfig
		setEnv map[string]string
		want   string
	}{
		{
			name: "test with no proxy env var",
			config: config.SupplierConfig{
				Key:     "tfstate",
				Backend: "gs",
				Path:    "terraform.tfstate",
			},
			setEnv: map[string]string{
				"AWS_DEFAULT_REGION": "us-east-1",
			},
			want: "us-east-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGSEnumerator(tt.config)
			assert.NotNil(t, got)
			assert.Equal(t, tt.config.Backend, got.config.Backend)
		})
	}
}

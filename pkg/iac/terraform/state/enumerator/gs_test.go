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
	}{
		{
			name: "test if required attributes are not nil",
			config: config.SupplierConfig{
				Key:     "tfstate",
				Backend: "gs",
				Path:    "terraform.tfstate",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGSEnumerator(tt.config)
			assert.NotNil(t, got)
			assert.NotNil(t, got.client)
			assert.NotNil(t, got.config)
		})
	}
}

func TestGSEnumerator_NewGSEnumerator_HasCorrectConfig(t *testing.T) {
	tests := []struct {
		name   string
		config config.SupplierConfig
	}{
		{
			name: "test whether the supplier config isn't changed",
			config: config.SupplierConfig{
				Key:     "tfstate",
				Backend: "gs",
				Path:    "terraform.tfstate",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewGSEnumerator(tt.config)
			assert.NotNil(t, got)
			assert.Equal(t, tt.config.Key, got.config.Key)
			assert.Equal(t, tt.config.Backend, got.config.Backend)
			assert.Equal(t, tt.config.Path, got.config.Path)
		})
	}
}

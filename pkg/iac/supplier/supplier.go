package supplier

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/iac/terraform/state/backend"

	"github.com/cloudskiff/driftctl/pkg/iac/config"

	"github.com/cloudskiff/driftctl/pkg/iac/terraform/state"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

var supportedSuppliers = []string{
	state.TerraformStateReaderSupplier,
}

func IsSupplierSupported(supplierKey string) bool {
	for _, s := range supportedSuppliers {
		if s == supplierKey {
			return true
		}
	}
	return false
}

func GetIACSupplier(config config.SupplierConfig) (resource.Supplier, error) {
	if !IsSupplierSupported(config.Key) {
		return nil, fmt.Errorf("Unsupported supplier '%s'", config.Key)
	}

	switch config.Key {
	case state.TerraformStateReaderSupplier:
		return state.NewReader(config)
	default:
		return nil, fmt.Errorf("Unsupported supplier '%s'", config.Key)
	}
}

func GetSupportedSuppliers() []string {
	return supportedSuppliers
}

func GetSupportedSchemes() []string {
	schemes := []string{
		"tfstate://",
	}
	for _, supplier := range supportedSuppliers {
		for _, backend := range backend.GetSupportedBackends() {
			schemes = append(schemes, fmt.Sprintf("%s+%s://", supplier, backend))
		}
	}
	return schemes
}

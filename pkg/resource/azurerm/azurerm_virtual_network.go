package azurerm

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AzureVirtualNetworkResourceType = "azurerm_virtual_network"

func initAzureVirtualNetworkMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzureVirtualNetworkResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)

		if v := res.Attributes().GetString("name"); v != nil && *v != "" {
			attrs["Name"] = *v
		}
		return attrs
	})
}

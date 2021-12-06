package azurerm

import "github.com/snyk/driftctl/pkg/resource"

const AzureResourceGroupResourceType = "azurerm_resource_group"

func initAzureResourceGroupMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzureResourceGroupResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if name := val.GetString("name"); name != nil && *name != "" {
			attrs["Name"] = *name
		}
		return attrs
	})
}

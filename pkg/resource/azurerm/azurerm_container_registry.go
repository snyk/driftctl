package azurerm

import "github.com/snyk/driftctl/enumeration/resource"

const AzureContainerRegistryResourceType = "azurerm_container_registry"

func initAzureContainerRegistryMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzureContainerRegistryResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if name := val.GetString("name"); name != nil && *name != "" {
			attrs["Name"] = *name
		}
		return attrs
	})
}

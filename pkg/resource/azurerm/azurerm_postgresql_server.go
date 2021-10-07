package azurerm

import "github.com/cloudskiff/driftctl/pkg/resource"

const AzurePostgresqlServerResourceType = "azurerm_postgresql_server"

func initAzurePostgresqlServerMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzurePostgresqlServerResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if name := val.GetString("name"); name != nil && *name != "" {
			attrs["Name"] = *name
		}
		return attrs
	})
}

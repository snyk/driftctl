package azurerm

import "github.com/snyk/driftctl/enumeration/resource"

const AzureRouteTableResourceType = "azurerm_route_table"

func initAzureRouteTableMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzureRouteTableResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)

		if v := res.Attributes().GetString("name"); v != nil && *v != "" {
			attrs["Name"] = *v
		}
		return attrs
	})
}

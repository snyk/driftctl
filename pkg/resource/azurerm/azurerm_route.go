package azurerm

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AzureRouteResourceType = "azurerm_route"

func initAzureRouteMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzureRouteResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)

		if v := res.Attributes().GetString("name"); v != nil && *v != "" {
			attrs["Name"] = *v
		}

		if v := res.Attributes().GetString("route_table_name"); v != nil && *v != "" {
			attrs["Table"] = *v
		}

		return attrs
	})
}

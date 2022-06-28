package azurerm

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AzureImageResourceType = "azurerm_image"

func initAzureImageMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzureImageResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)

		if v := res.Attributes().GetString("name"); v != nil && *v != "" {
			attrs["Name"] = *v
		}

		return attrs
	})
}

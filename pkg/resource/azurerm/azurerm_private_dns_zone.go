package azurerm

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/azurerm"
)

func initAzurePrivateDNSZoneMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(azurerm.AzurePrivateDNSZoneResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"number_of_record_sets"})
		res.Attributes().SafeDelete([]string{"timeouts"})
	})
}

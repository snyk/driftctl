package azurerm

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AzurePrivateDNSZoneResourceType = "azurerm_private_dns_zone"

func initAzurePrivateDNSZoneMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AzurePrivateDNSZoneResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"number_of_record_sets"})
		res.Attributes().SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetFlags(AzurePrivateDNSZoneResourceType, resource.FlagDeepMode)
}

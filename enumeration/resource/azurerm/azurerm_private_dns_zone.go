package azurerm

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AzurePrivateDNSZoneResourceType = "azurerm_private_dns_zone"

func initAzurePrivateDNSZoneMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AzurePrivateDNSZoneResourceType, resource.FlagDeepMode)
}

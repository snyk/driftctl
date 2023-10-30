package azurerm

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AzurePrivateDNSSRVRecordResourceType = "azurerm_private_dns_srv_record"

func initAzurePrivateDNSSRVRecordMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AzurePrivateDNSSRVRecordResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzurePrivateDNSSRVRecordResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if name := val.GetString("name"); name != nil && *name != "" {
			attrs["Name"] = *name
		}
		if zone := val.GetString("zone_name"); zone != nil && *zone != "" {
			attrs["Zone"] = *zone
		}
		return attrs
	})
}

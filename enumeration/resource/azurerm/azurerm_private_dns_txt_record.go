package azurerm

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AzurePrivateDNSTXTRecordResourceType = "azurerm_private_dns_txt_record"

func initAzurePrivateDNSTXTRecordMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzurePrivateDNSTXTRecordResourceType, func(res *resource.Resource) map[string]string {
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
	resourceSchemaRepository.SetFlags(AzurePrivateDNSTXTRecordResourceType, resource.FlagDeepMode)
}

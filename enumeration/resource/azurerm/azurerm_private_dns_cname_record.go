package azurerm

import "github.com/snyk/driftctl/enumeration/resource"

const AzurePrivateDNSCNameRecordResourceType = "azurerm_private_dns_cname_record"

func initAzurePrivateDNSCNameRecordMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AzurePrivateDNSCNameRecordResourceType, resource.FlagDeepMode)

	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzurePrivateDNSCNameRecordResourceType, func(res *resource.Resource) map[string]string {
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

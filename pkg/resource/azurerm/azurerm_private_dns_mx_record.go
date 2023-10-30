package azurerm

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AzurePrivateDNSMXRecordResourceType = "azurerm_private_dns_mx_record"

func initAzurePrivateDNSMXRecordMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AzurePrivateDNSMXRecordResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetNormalizeFunc(AzurePrivateDNSMXRecordResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzurePrivateDNSMXRecordResourceType, func(res *resource.Resource) map[string]string {
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

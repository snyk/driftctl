package azurerm

import "github.com/snyk/driftctl/enumeration/resource"

const AzureNetworkSecurityGroupResourceType = "azurerm_network_security_group"

func initAzureNetworkSecurityGroupMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AzureNetworkSecurityGroupResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzureNetworkSecurityGroupResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if name := val.GetString("name"); name != nil && *name != "" {
			attrs["Name"] = *name
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(AzureNetworkSecurityGroupResourceType, resource.FlagDeepMode)
}

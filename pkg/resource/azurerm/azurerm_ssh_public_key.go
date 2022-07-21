package azurerm

import "github.com/snyk/driftctl/enumeration/resource"

const AzureSSHPublicKeyResourceType = "azurerm_ssh_public_key"

func initAzureSSHPublicKeyMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AzureSSHPublicKeyResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzureSSHPublicKeyResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)

		if v := res.Attributes().GetString("name"); v != nil && *v != "" {
			attrs["Name"] = *v
		}

		return attrs
	})
	resourceSchemaRepository.SetFlags(AzureSSHPublicKeyResourceType, resource.FlagDeepMode)
}

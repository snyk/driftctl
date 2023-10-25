package azurerm

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AzureSSHPublicKeyResourceType = "azurerm_ssh_public_key"

func initAzureSSHPublicKeyMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
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
}

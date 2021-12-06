package azurerm

import "github.com/snyk/driftctl/pkg/resource"

const AzureLoadBalancerResourceType = "azurerm_lb"

func initAzureLoadBalancerMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzureLoadBalancerResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if name := val.GetString("name"); name != nil && *name != "" {
			attrs["Name"] = *name
		}
		return attrs
	})
}

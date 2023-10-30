package azurerm

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AzureLoadBalancerRuleResourceType = "azurerm_lb_rule"

func initAzureLoadBalancerRuleMetadata(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AzureLoadBalancerRuleResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AzureLoadBalancerRuleResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)
		if name := res.Attributes().GetString("name"); name != nil && *name != "" {
			attrs["Name"] = *name
		}
		return attrs
	})
}

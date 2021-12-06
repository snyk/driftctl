package middlewares

import (
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/azurerm"
)

// Explodes subnet found in azurerm_virtual_network.subnet from state resources to dedicated resources
type AzurermSubnetExpander struct {
	resourceFactory resource.ResourceFactory
}

func NewAzurermSubnetExpander(resourceFactory resource.ResourceFactory) AzurermSubnetExpander {
	return AzurermSubnetExpander{
		resourceFactory: resourceFactory,
	}
}

func (m AzurermSubnetExpander) Execute(_, resourcesFromState *[]*resource.Resource) error {
	newList := make([]*resource.Resource, 0)
	for _, res := range *resourcesFromState {
		newList = append(newList, res)

		// Ignore all resources other than azurerm_virtual_network
		if res.ResourceType() != azurerm.AzureVirtualNetworkResourceType {
			continue
		}

		subnets, exist := res.Attributes().Get("subnet")
		if !exist || subnets == nil {
			continue
		}

		for _, subnet := range subnets.([]interface{}) {
			subnet := subnet.(map[string]interface{})
			id := subnet["id"].(string)
			exist := false
			for _, resFromState := range *resourcesFromState {
				if resFromState.ResourceType() == azurerm.AzureSubnetResourceType &&
					resFromState.ResourceId() == id {
					exist = true
					break
				}
			}
			if exist {
				continue
			}
			res := m.resourceFactory.CreateAbstractResource(
				azurerm.AzureSubnetResourceType,
				id,
				map[string]interface{}{},
			)

			newList = append(newList, res)

		}

		res.Attrs.SafeDelete([]string{"subnet"})
	}
	*resourcesFromState = newList
	return nil
}

package middlewares

import (
	"strings"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/azurerm"
)

// Explodes routes found in azurerm_route_table.route from state resources to dedicated resources
type AzurermRouteExpander struct {
	resourceFactory resource.ResourceFactory
}

func NewAzurermRouteExpander(resourceFactory resource.ResourceFactory) AzurermRouteExpander {
	return AzurermRouteExpander{
		resourceFactory: resourceFactory,
	}
}

func (m AzurermRouteExpander) Execute(_, resourcesFromState *[]*resource.Resource) error {
	newList := make([]*resource.Resource, 0)
	for _, res := range *resourcesFromState {

		newList = append(newList, res)

		// Ignore all resources other than route tables
		if res.ResourceType() != azurerm.AzureRouteTableResourceType {
			continue
		}

		routes, exist := res.Attributes().Get("route")
		if !exist || routes == nil {
			continue
		}

		for _, route := range routes.([]interface{}) {
			route := route.(map[string]interface{})
			id := strings.Join([]string{res.ResourceId(), "routes", route["name"].(string)}, "/")
			exist := false
			for _, resFromState := range *resourcesFromState {
				if resFromState.ResourceType() == azurerm.AzureRouteResourceType &&
					resFromState.ResourceId() == id {
					exist = true
					break
				}
			}
			if exist {
				continue
			}
			expandedRoute := m.resourceFactory.CreateAbstractResource(
				azurerm.AzureRouteResourceType,
				id,
				map[string]interface{}{
					"name":             route["name"].(string),
					"route_table_name": *res.Attributes().GetString("name"),
				},
			)
			newList = append(newList, expandedRoute)
		}

		res.Attributes().SafeDelete([]string{"route"})
	}
	*resourcesFromState = newList
	return nil
}

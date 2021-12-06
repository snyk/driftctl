package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// Explodes api gateway default resource found in aws_api_gateway_rest_api.root_resource_id from state resources to dedicated resources
type AwsApiGatewayResourceExpander struct {
	resourceFactory resource.ResourceFactory
}

func NewAwsApiGatewayResourceExpander(resourceFactory resource.ResourceFactory) AwsApiGatewayResourceExpander {
	return AwsApiGatewayResourceExpander{
		resourceFactory: resourceFactory,
	}
}

func (m AwsApiGatewayResourceExpander) Execute(_, resourcesFromState *[]*resource.Resource) error {
	newStateResources := make([]*resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than aws_api_gateway_rest_api
		if res.ResourceType() != aws.AwsApiGatewayRestApiResourceType {
			newStateResources = append(newStateResources, res)
			continue
		}

		newStateResources = append(newStateResources, res)

		err := m.handleResource(res, &newStateResources)
		if err != nil {
			return err
		}
	}
	*resourcesFromState = newStateResources
	return nil
}

func (m *AwsApiGatewayResourceExpander) handleResource(api *resource.Resource, results *[]*resource.Resource) error {
	resourceId := api.Attrs.GetString("root_resource_id")
	if resourceId == nil || *resourceId == "" {
		return nil
	}

	newResource := m.resourceFactory.CreateAbstractResource(aws.AwsApiGatewayResourceResourceType, *resourceId, map[string]interface{}{
		"rest_api_id": api.ResourceId(),
		"path":        "/",
	})

	*results = append(*results, newResource)
	logrus.WithFields(logrus.Fields{
		"id": newResource.ResourceId(),
	}).Debug("Created new resource from api gateway rest api")

	return nil
}

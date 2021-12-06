package middlewares

import (
	"strings"

	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// Create a aws_api_gateway_stage resource from a aws_api_gateway_deployment resource and ignore the latter resource
// since we don't support it
type AwsApiGatewayDeploymentExpander struct {
	resourceFactory resource.ResourceFactory
}

func NewAwsApiGatewayDeploymentExpander(resourceFactory resource.ResourceFactory) AwsApiGatewayDeploymentExpander {
	return AwsApiGatewayDeploymentExpander{resourceFactory}
}

func (m AwsApiGatewayDeploymentExpander) Execute(_, resourcesFromState *[]*resource.Resource) error {
	var newResources []*resource.Resource
	for _, res := range *resourcesFromState {
		if res.ResourceType() != aws.AwsApiGatewayDeploymentResourceType {
			newResources = append(newResources, res)
			continue
		}

		stageName := res.Attributes().GetString("stage_name")
		if stageName == nil || *stageName == "" {
			continue
		}

		newStage := m.resourceFactory.CreateAbstractResource(
			aws.AwsApiGatewayStageResourceType,
			strings.Join([]string{"ags", *(res.Attributes().GetString("rest_api_id")), *stageName}, "-"),
			map[string]interface{}{},
		)

		newResources = append(newResources, newStage)
	}
	*resourcesFromState = newResources

	return nil
}

package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// Explodes policy found in aws_api_gateway_rest_api.policy from state resources to dedicated resources
type AwsApiGatewayRestApiPolicyExpander struct {
	resourceFactory resource.ResourceFactory
}

func NewAwsApiGatewayRestApiPolicyExpander(resourceFactory resource.ResourceFactory) AwsApiGatewayRestApiPolicyExpander {
	return AwsApiGatewayRestApiPolicyExpander{
		resourceFactory: resourceFactory,
	}
}

func (m AwsApiGatewayRestApiPolicyExpander) Execute(_, resourcesFromState *[]*resource.Resource) error {
	newList := make([]*resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than api_gateway_rest_api
		if res.ResourceType() != aws.AwsApiGatewayRestApiResourceType {
			newList = append(newList, res)
			continue
		}

		newList = append(newList, res)

		if hasRestApiPolicyAttached(res.ResourceId(), resourcesFromState) {
			res.Attrs.SafeDelete([]string{"policy"})
			continue
		}

		err := m.handlePolicy(res, &newList)
		if err != nil {
			return err
		}
	}
	*resourcesFromState = newList
	return nil
}

func (m *AwsApiGatewayRestApiPolicyExpander) handlePolicy(api *resource.Resource, results *[]*resource.Resource) error {
	policy, exist := api.Attrs.Get("policy")
	if !exist || policy == nil || policy == "" {
		return nil
	}

	data := map[string]interface{}{
		"id":          api.ResourceId(),
		"rest_api_id": api.ResourceId(),
		"policy":      policy,
	}

	newPolicy := m.resourceFactory.CreateAbstractResource(aws.AwsApiGatewayRestApiPolicyResourceType, api.ResourceId(), data)

	*results = append(*results, newPolicy)
	logrus.WithFields(logrus.Fields{
		"id": newPolicy.ResourceId(),
	}).Debug("Created new policy from api gateway rest api")

	api.Attrs.SafeDelete([]string{"policy"})
	return nil
}

// Return true if the rest api has a aws_api_gateway_rest_api_policy resource attached to itself.
// It is mandatory since it's possible to have a aws_api_gateway_rest_api with an inline policy
// AND a aws_api_gateway_rest_api_policy resource at the same time. At the end, on the AWS console,
// the aws_api_gateway_rest_api_policy will be used.
func hasRestApiPolicyAttached(api string, resourcesFromState *[]*resource.Resource) bool {
	for _, res := range *resourcesFromState {
		if res.ResourceType() == aws.AwsApiGatewayRestApiPolicyResourceType &&
			res.ResourceId() == api {
			return true
		}
	}
	return false
}

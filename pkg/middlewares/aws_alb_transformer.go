package middlewares

import (
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsALBTransformer is a simple middleware to turn all aws_alb resources into aws_lb ones
// Both types provide the same functionality, but we can't know which one was used to provision cloud resources.
// So we use aws_lb as the common type.
type AwsALBTransformer struct {
	resourceFactory resource.ResourceFactory
}

func NewAwsALBTransformer(resourceFactory resource.ResourceFactory) AwsALBTransformer {
	return AwsALBTransformer{
		resourceFactory: resourceFactory,
	}
}

func (m AwsALBTransformer) Execute(_, resourcesFromState *[]*resource.Resource) error {
	newStateResources := make([]*resource.Resource, 0, len(*resourcesFromState))

	for _, res := range *resourcesFromState {
		if res.ResourceType() != aws.AwsApplicationLoadBalancerResourceType {
			newStateResources = append(newStateResources, res)
			continue
		}

		newStateResources = append(newStateResources, m.resourceFactory.CreateAbstractResource(
			aws.AwsLoadBalancerResourceType,
			res.ResourceId(),
			*res.Attributes(),
		))
	}

	*resourcesFromState = newStateResources
	return nil
}

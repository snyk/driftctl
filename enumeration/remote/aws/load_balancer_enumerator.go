package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type LoadBalancerEnumerator struct {
	repository repository.ELBV2Repository
	factory    resource.ResourceFactory
}

func NewLoadBalancerEnumerator(repo repository.ELBV2Repository, factory resource.ResourceFactory) *LoadBalancerEnumerator {
	return &LoadBalancerEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *LoadBalancerEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsLoadBalancerResourceType
}

func (e *LoadBalancerEnumerator) Enumerate() ([]*resource.Resource, error) {
	loadBalancers, err := e.repository.ListAllLoadBalancers()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(loadBalancers))

	for _, lb := range loadBalancers {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*lb.LoadBalancerArn,
				map[string]interface{}{
					"name": *lb.LoadBalancerName,
				},
			),
		)
	}

	return results, err
}

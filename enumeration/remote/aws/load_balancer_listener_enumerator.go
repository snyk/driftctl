package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type LoadBalancerListenerEnumerator struct {
	repository repository.ELBV2Repository
	factory    resource.ResourceFactory
}

func NewLoadBalancerListenerEnumerator(repo repository.ELBV2Repository, factory resource.ResourceFactory) *LoadBalancerListenerEnumerator {
	return &LoadBalancerListenerEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *LoadBalancerListenerEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsLoadBalancerListenerResourceType
}

func (e *LoadBalancerListenerEnumerator) Enumerate() ([]*resource.Resource, error) {
	loadBalancers, err := e.repository.ListAllLoadBalancers()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsLoadBalancerResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, lb := range loadBalancers {
		listeners, err := e.repository.ListAllLoadBalancerListeners(*lb.LoadBalancerArn)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, listener := range listeners {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*listener.ListenerArn,
					map[string]interface{}{},
				),
			)
		}
	}

	return results, nil
}

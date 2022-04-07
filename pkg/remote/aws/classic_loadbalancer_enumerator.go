package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type ClassicLoadBalancerEnumerator struct {
	repository repository.ELBRepository
	factory    resource.ResourceFactory
}

func NewClassicLoadBalancerEnumerator(repo repository.ELBRepository, factory resource.ResourceFactory) *ClassicLoadBalancerEnumerator {
	return &ClassicLoadBalancerEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ClassicLoadBalancerEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsClassicLoadBalancerResourceType
}

func (e *ClassicLoadBalancerEnumerator) Enumerate() ([]*resource.Resource, error) {
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
				*lb.LoadBalancerName,
				map[string]interface{}{},
			),
		)
	}

	return results, nil
}

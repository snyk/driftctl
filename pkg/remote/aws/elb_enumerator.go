package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

type ElbEnumerator struct {
	repo    repository.ELBRepository
	factory resource.ResourceFactory
}

func NewElbEnumerator(repo repository.ELBRepository, factory resource.ResourceFactory) *ElbEnumerator {
	return &ElbEnumerator{
		repo,
		factory,
	}
}

func (e *ElbEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsElbResourceType
}

func (e *ElbEnumerator) Enumerate() ([]resource.Resource, error) {
	loadBalancers, err := e.repo.ListLoadBalancers()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}
	results := make([]resource.Resource, 0)

	for _, item := range loadBalancers {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*item.LoadBalancerName,
				map[string]interface{}{},
			),
		)
	}

	return results, nil
}

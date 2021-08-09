package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

type ElbEnumerator struct {
	repo    repository.AutoScalingRepository
	factory resource.ResourceFactory
}

func NewElbEnumerator(repo repository.AutoScalingRepository, factory resource.ResourceFactory) *ElbEnumerator {
	return &ElbEnumerator{
		repo,
		factory,
	}
}

func (e *ElbEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsElbResourceType
}

func (e *ElbEnumerator) Enumerate() ([]resource.Resource, error) {
	groups, err := e.repo.DescribeGroups([]*string{})
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsAutoScalingGroupResourceType)
	}

	results := make([]resource.Resource, 0)

	for _, group := range groups {
		loadBalancers, err := e.repo.DescribeLoadBalancers(*group.AutoScalingGroupName)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

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
	}

	return results, nil
}

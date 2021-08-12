package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

type AutoScalingGroupsEnumerator struct {
	repo    repository.AutoScalingRepository
	factory resource.ResourceFactory
}

func NewAutoScalingGroupsEnumerator(repo repository.AutoScalingRepository, factory resource.ResourceFactory) *AutoScalingGroupsEnumerator {
	return &AutoScalingGroupsEnumerator{
		repo,
		factory,
	}
}

func (e *AutoScalingGroupsEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsAutoScalingGroupResourceType
}

func (e *AutoScalingGroupsEnumerator) Enumerate() ([]resource.Resource, error) {
	groups, err := e.repo.ListGroups([]*string{})
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, 0, len(groups))

	for _, item := range groups {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*item.AutoScalingGroupName,
				map[string]interface{}{},
			),
		)
	}

	return results, nil
}

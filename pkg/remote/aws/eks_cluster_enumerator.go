package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

type EKSClusterEnumerator struct {
	repo    repository.EKSRepository
	factory resource.ResourceFactory
}

func NewEKSClusterEnumerator(repo repository.EKSRepository, factory resource.ResourceFactory) *EKSClusterEnumerator {
	return &EKSClusterEnumerator{
		repo,
		factory,
	}
}

func (e *EKSClusterEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsEKSClusterResourceType
}

func (e *EKSClusterEnumerator) Enumerate() ([]resource.Resource, error) {
	clusters, err := e.repo.ListAllClusters()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, 0, len(clusters))

	for _, item := range clusters {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*item,
				map[string]interface{}{},
			),
		)
	}

	return results, nil
}

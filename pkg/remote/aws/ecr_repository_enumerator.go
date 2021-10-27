package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type ECRRepositoryEnumerator struct {
	repository repository.ECRRepository
	factory    resource.ResourceFactory
}

func NewECRRepositoryEnumerator(repo repository.ECRRepository, factory resource.ResourceFactory) *ECRRepositoryEnumerator {
	return &ECRRepositoryEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ECRRepositoryEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsEcrRepositoryResourceType
}

func (e *ECRRepositoryEnumerator) Enumerate() ([]*resource.Resource, error) {
	repos, err := e.repository.ListAllRepositories()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(repos))

	for _, repo := range repos {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*repo.RepositoryName,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

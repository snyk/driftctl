package aws

import (
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type ECRRepositoryPolicyEnumerator struct {
	repository repository.ECRRepository
	factory    resource.ResourceFactory
}

func NewECRRepositoryPolicyEnumerator(repo repository.ECRRepository, factory resource.ResourceFactory) *ECRRepositoryPolicyEnumerator {
	return &ECRRepositoryPolicyEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ECRRepositoryPolicyEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsEcrRepositoryPolicyResourceType
}

func (e *ECRRepositoryPolicyEnumerator) Enumerate() ([]*resource.Resource, error) {
	repos, err := e.repository.ListAllRepositories()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsEcrRepositoryResourceType)
	}

	results := make([]*resource.Resource, 0, len(repos))

	for _, repo := range repos {
		repoOutput, err := e.repository.GetRepositoryPolicy(repo)
		if _, ok := err.(*ecr.RepositoryPolicyNotFoundException); ok {
			continue
		}
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*repoOutput.RepositoryName,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

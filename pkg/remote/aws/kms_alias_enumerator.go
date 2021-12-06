package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type KMSAliasEnumerator struct {
	repository repository.KMSRepository
	factory    resource.ResourceFactory
}

func NewKMSAliasEnumerator(repo repository.KMSRepository, factory resource.ResourceFactory) *KMSAliasEnumerator {
	return &KMSAliasEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *KMSAliasEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsKmsAliasResourceType
}

func (e *KMSAliasEnumerator) Enumerate() ([]*resource.Resource, error) {
	aliases, err := e.repository.ListAllAliases()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(aliases))

	for _, alias := range aliases {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*alias.AliasName,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

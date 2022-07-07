package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type KMSKeyEnumerator struct {
	repository repository.KMSRepository
	factory    resource.ResourceFactory
}

func NewKMSKeyEnumerator(repo repository.KMSRepository, factory resource.ResourceFactory) *KMSKeyEnumerator {
	return &KMSKeyEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *KMSKeyEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsKmsKeyResourceType
}

func (e *KMSKeyEnumerator) Enumerate() ([]*resource.Resource, error) {
	keys, err := e.repository.ListAllKeys()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(keys))

	for _, key := range keys {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*key.KeyId,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

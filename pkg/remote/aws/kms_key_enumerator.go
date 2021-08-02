package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
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

func (e *KMSKeyEnumerator) Enumerate() ([]resource.Resource, error) {
	keys, err := e.repository.ListAllKeys()
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(keys))

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

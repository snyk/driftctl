package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type EC2KeyPairEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2KeyPairEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2KeyPairEnumerator {
	return &EC2KeyPairEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2KeyPairEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsKeyPairResourceType
}

func (e *EC2KeyPairEnumerator) Enumerate() ([]*resource.Resource, error) {
	keyPairs, err := e.repository.ListAllKeyPairs()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(keyPairs))

	for _, keyPair := range keyPairs {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*keyPair.KeyName,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

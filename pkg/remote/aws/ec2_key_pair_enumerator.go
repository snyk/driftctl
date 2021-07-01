package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	tf "github.com/cloudskiff/driftctl/pkg/remote/terraform"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type EC2KeyPairEnumerator struct {
	repository     repository.EC2Repository
	factory        resource.ResourceFactory
	providerConfig tf.TerraformProviderConfig
}

func NewEC2KeyPairEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory, providerConfig tf.TerraformProviderConfig) *EC2KeyPairEnumerator {
	return &EC2KeyPairEnumerator{
		repository:     repo,
		factory:        factory,
		providerConfig: providerConfig,
	}
}

func (e *EC2KeyPairEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsKeyPairResourceType
}

func (e *EC2KeyPairEnumerator) Enumerate() ([]resource.Resource, error) {
	keyPairs, err := e.repository.ListAllKeyPairs()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(keyPairs))

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

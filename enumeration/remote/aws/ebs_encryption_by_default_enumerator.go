package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type EC2EbsEncryptionByDefaultEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2EbsEncryptionByDefaultEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2EbsEncryptionByDefaultEnumerator {
	return &EC2EbsEncryptionByDefaultEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2EbsEncryptionByDefaultEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsEbsEncryptionByDefaultResourceType
}

func (e *EC2EbsEncryptionByDefaultEnumerator) Enumerate() ([]*resource.Resource, error) {
	enabled, err := e.repository.IsEbsEncryptionEnabledByDefault()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0)

	results = append(
		results,
		e.factory.CreateAbstractResource(
			string(e.SupportedType()),
			"ebs_encryption_default",
			map[string]interface{}{
				"enabled": enabled,
			},
		),
	)

	return results, err
}

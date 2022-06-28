package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type EC2AmiEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2AmiEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2AmiEnumerator {
	return &EC2AmiEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2AmiEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsAmiResourceType
}

func (e *EC2AmiEnumerator) Enumerate() ([]*resource.Resource, error) {
	images, err := e.repository.ListAllImages()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(images))

	for _, image := range images {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*image.ImageId,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

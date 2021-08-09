package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
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

	results := make([]*resource.Resource, len(images))

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

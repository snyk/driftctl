package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type EC2EbsVolumeEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2EbsVolumeEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2EbsVolumeEnumerator {
	return &EC2EbsVolumeEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2EbsVolumeEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsEbsVolumeResourceType
}

func (e *EC2EbsVolumeEnumerator) Enumerate() ([]*resource.Resource, error) {
	volumes, err := e.repository.ListAllVolumes()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(volumes))

	for _, volume := range volumes {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*volume.VolumeId,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
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

func (e *EC2EbsVolumeEnumerator) Enumerate() ([]resource.Resource, error) {
	volumes, err := e.repository.ListAllVolumes()
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(volumes))

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

package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	tf "github.com/cloudskiff/driftctl/pkg/remote/terraform"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type EC2EbsVolumeEnumerator struct {
	repository     repository.EC2Repository
	factory        resource.ResourceFactory
	providerConfig tf.TerraformProviderConfig
}

func NewEC2EbsVolumeEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory, providerConfig tf.TerraformProviderConfig) *EC2EbsVolumeEnumerator {
	return &EC2EbsVolumeEnumerator{
		repository:     repo,
		factory:        factory,
		providerConfig: providerConfig,
	}
}

func (e *EC2EbsVolumeEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsEbsVolumeResourceType
}

func (e *EC2EbsVolumeEnumerator) Enumerate() ([]resource.Resource, error) {
	volumes, err := e.repository.ListAllVolumes()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, string(e.SupportedType()))
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

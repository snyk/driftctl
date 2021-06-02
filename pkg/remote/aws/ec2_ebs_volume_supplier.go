package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type EC2EbsVolumeSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.EC2Repository
	runner       *terraform.ParallelResourceReader
}

func NewEC2EbsVolumeSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer) *EC2EbsVolumeSupplier {
	return &EC2EbsVolumeSupplier{
		provider,
		deserializer,
		repository.NewEC2Repository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *EC2EbsVolumeSupplier) Resources() ([]resource.Resource, error) {
	volumes, err := s.client.ListAllVolumes()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsEbsVolumeResourceType)
	}
	results := make([]cty.Value, 0)
	if len(volumes) > 0 {
		for _, volume := range volumes {
			vol := *volume
			s.runner.Run(func() (cty.Value, error) {
				return s.readEbsVolume(vol)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(resourceaws.AwsEbsVolumeResourceType, results)
}

func (s *EC2EbsVolumeSupplier) readEbsVolume(volume ec2.Volume) (cty.Value, error) {
	id := aws.StringValue(volume.VolumeId)
	resVolume, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: resourceaws.AwsEbsVolumeResourceType,
		ID: id,
	})
	if err != nil {
		logrus.Warnf("Error reading volume %s[%s]: %+v", id, resourceaws.AwsEbsVolumeResourceType, err)
		return cty.NilVal, err
	}
	return *resVolume, nil
}

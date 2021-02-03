package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type EC2EbsVolumeSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       ec2iface.EC2API
	runner       *terraform.ParallelResourceReader
}

func NewEC2EbsVolumeSupplier(provider *TerraformProvider) *EC2EbsVolumeSupplier {
	return &EC2EbsVolumeSupplier{
		provider,
		awsdeserializer.NewEC2EbsVolumeDeserializer(),
		ec2.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s EC2EbsVolumeSupplier) Resources() ([]resource.Resource, error) {
	volumes, err := listVolumes(s.client)
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
	return s.deserializer.Deserialize(results)
}

func (s EC2EbsVolumeSupplier) readEbsVolume(volume ec2.Volume) (cty.Value, error) {
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

func listVolumes(client ec2iface.EC2API) ([]*ec2.Volume, error) {
	var volumes []*ec2.Volume
	input := &ec2.DescribeVolumesInput{}
	err := client.DescribeVolumesPages(input, func(res *ec2.DescribeVolumesOutput, lastPage bool) bool {
		volumes = append(volumes, res.Volumes...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return volumes, nil
}

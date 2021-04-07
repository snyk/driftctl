package middlewares

import (
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// Remove root_block_device from aws_instance resources and create dedicated aws_ebs_volume resources
type AwsInstanceBlockDeviceResourceMapper struct {
	resourceFactory resource.ResourceFactory
}

func NewAwsInstanceBlockDeviceResourceMapper(resourceFactory resource.ResourceFactory) AwsInstanceBlockDeviceResourceMapper {
	return AwsInstanceBlockDeviceResourceMapper{resourceFactory: resourceFactory}
}

func (a AwsInstanceBlockDeviceResourceMapper) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {

	newStateResources := make([]resource.Resource, 0)
	for _, stateRes := range *resourcesFromState {

		// Ignore all resources other than aws_instance
		if stateRes.TerraformType() != aws.AwsInstanceResourceType {
			newStateResources = append(newStateResources, stateRes)
			continue
		}

		instance, _ := stateRes.(*aws.AwsInstance)
		if instance.RootBlockDevice != nil && len(*instance.RootBlockDevice) > 0 {
			for _, rootBlock := range *instance.RootBlockDevice {
				logrus.WithFields(logrus.Fields{
					"volume":   *rootBlock.VolumeId,
					"instance": instance.TerraformId(),
				}).Debug("Creating aws_ebs_volume from aws_instance.root_block_device")
				data := map[string]interface{}{
					"availability_zone":    instance.AvailabilityZone,
					"encrypted":            rootBlock.Encrypted,
					"id":                   *rootBlock.VolumeId,
					"iops":                 rootBlock.Iops,
					"kms_key_id":           rootBlock.KmsKeyId,
					"size":                 rootBlock.VolumeSize,
					"type":                 rootBlock.VolumeType,
					"multi_attach_enabled": false,
					"tags":                 instance.VolumeTags,
				}
				ctyVal, err := a.resourceFactory.CreateResource(data, "aws_ebs_volume")
				if err != nil {
					return err
				}
				ebsVolume := aws.AwsEbsVolume{
					AvailabilityZone:   instance.AvailabilityZone,
					Encrypted:          rootBlock.Encrypted,
					Id:                 *rootBlock.VolumeId,
					Iops:               rootBlock.Iops,
					KmsKeyId:           rootBlock.KmsKeyId,
					Size:               rootBlock.VolumeSize,
					Type:               rootBlock.VolumeType,
					MultiAttachEnabled: awssdk.Bool(false),
					Tags:               instance.VolumeTags,
					CtyVal:             ctyVal,
				}
				newStateResources = append(newStateResources, &ebsVolume)
			}
			instance.RootBlockDevice = nil
		}
		if instance.EbsBlockDevice != nil && len(*instance.EbsBlockDevice) > 0 {
			for _, blockDevice := range *instance.EbsBlockDevice {
				logrus.WithFields(logrus.Fields{
					"volume":   *blockDevice.VolumeId,
					"instance": instance.TerraformId(),
				}).Debug("Creating aws_ebs_volume from aws_instance.ebs_block_device")
				data := map[string]interface{}{
					"availability_zone":    instance.AvailabilityZone,
					"encrypted":            blockDevice.Encrypted,
					"id":                   *blockDevice.VolumeId,
					"iops":                 blockDevice.Iops,
					"kms_key_id":           blockDevice.KmsKeyId,
					"size":                 blockDevice.VolumeSize,
					"type":                 blockDevice.VolumeType,
					"multi_attach_enabled": false,
					"tags":                 instance.VolumeTags,
				}
				ctyVal, err := a.resourceFactory.CreateResource(data, "aws_ebs_volume")
				if err != nil {
					return err
				}
				ebsVolume := aws.AwsEbsVolume{
					AvailabilityZone:   instance.AvailabilityZone,
					Encrypted:          blockDevice.Encrypted,
					Id:                 *blockDevice.VolumeId,
					Iops:               blockDevice.Iops,
					KmsKeyId:           blockDevice.KmsKeyId,
					Size:               blockDevice.VolumeSize,
					Type:               blockDevice.VolumeType,
					MultiAttachEnabled: awssdk.Bool(false),
					Tags:               instance.VolumeTags,
					CtyVal:             ctyVal,
				}
				newStateResources = append(newStateResources, &ebsVolume)
			}
			instance.EbsBlockDevice = nil
		}
		newStateResources = append(newStateResources, instance)
	}

	*resourcesFromState = newStateResources

	return nil
}

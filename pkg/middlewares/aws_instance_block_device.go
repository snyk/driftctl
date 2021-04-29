package middlewares

import (
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// Remove root_block_device from aws_instance resources and create dedicated aws_ebs_volume resources
type AwsInstanceBlockDeviceResourceMapper struct {
	resourceFactory          resource.ResourceFactory
	resourceSchemaRepository resource.SchemaRepositoryInterface
}

func NewAwsInstanceBlockDeviceResourceMapper(resourceFactory resource.ResourceFactory, resourceSchemaRepository resource.SchemaRepositoryInterface) AwsInstanceBlockDeviceResourceMapper {
	return AwsInstanceBlockDeviceResourceMapper{
		resourceFactory:          resourceFactory,
		resourceSchemaRepository: resourceSchemaRepository,
	}
}

func (a AwsInstanceBlockDeviceResourceMapper) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {

	newStateResources := make([]resource.Resource, 0)
	for _, stateRes := range *resourcesFromState {

		// Ignore all resources other than aws_instance
		if stateRes.TerraformType() != aws.AwsInstanceResourceType {
			newStateResources = append(newStateResources, stateRes)
			continue
		}

		if resource.IsRefactoredResource("aws_instance") {
			instance, _ := stateRes.(*resource.AbstractResource)

			if rootBlockDevice, exist := instance.Attrs.Get("root_block_device"); exist {
				for _, rootBlock := range rootBlockDevice.([]interface{}) {
					rootBlock := rootBlock.(map[string]interface{})
					logrus.WithFields(logrus.Fields{
						"volume":   rootBlock["volume_id"],
						"instance": instance.TerraformId(),
					}).Debug("Creating aws_ebs_volume from aws_instance.root_block_device")
					data := map[string]interface{}{
						"availability_zone": func() interface{} {
							if val, ok := instance.Attrs.Get("availability_zone"); ok {
								return val
							}
							return nil
						}(),
						"encrypted":            rootBlock["encrypted"],
						"id":                   rootBlock["volume_id"],
						"iops":                 rootBlock["iops"],
						"kms_key_id":           rootBlock["kms_key_id"],
						"size":                 rootBlock["volume_size"],
						"type":                 rootBlock["volume_type"],
						"multi_attach_enabled": false,
						"tags": func() interface{} {
							if val, ok := instance.Attrs.Get("volume_tags"); ok {
								return val
							}
							return nil
						}(),
					}
					ctyVal, err := a.resourceFactory.CreateResource(data, "aws_ebs_volume")
					if err != nil {
						return err
					}
					ebsVolume := aws.AwsEbsVolume{
						AvailabilityZone: func() *string {
							if val, ok := instance.Attrs.Get("availability_zone"); ok {
								return awssdk.String(val.(string))
							}
							return nil
						}(),
						Encrypted: func() *bool {
							if val, ok := rootBlock["encrypted"]; ok {
								return awssdk.Bool(val.(bool))
							}
							return nil
						}(),
						Id: rootBlock["volume_id"].(string),
						Iops: func() *int {
							if val, ok := rootBlock["iops"]; ok {
								return awssdk.Int(int(val.(float64)))
							}
							return nil
						}(),
						KmsKeyId: func() *string {
							if val, ok := rootBlock["kms_key_id"]; ok {
								return awssdk.String(val.(string))
							}
							return nil
						}(),
						Size: func() *int {
							if val, ok := rootBlock["volume_size"]; ok {
								return awssdk.Int(int(val.(float64)))
							}
							return nil
						}(),
						Type: func() *string {
							if val, ok := rootBlock["volume_type"]; ok {
								return awssdk.String(val.(string))
							}
							return nil
						}(),
						MultiAttachEnabled: awssdk.Bool(false),
						Tags: func() map[string]string {
							if val, ok := instance.Attrs.Get("volume_tags"); ok {
								return val.(map[string]string)
							}
							return nil
						}(),
						CtyVal: ctyVal,
					}
					newStateResources = append(newStateResources, &ebsVolume)
				}
				instance.Attrs.SafeDelete([]string{"root_block_device"})
			}

			if ebsBlockDevice, exist := instance.Attrs.Get("ebs_block_device"); exist {
				for _, blockDevice := range ebsBlockDevice.([]interface{}) {
					blockDevice := blockDevice.(map[string]interface{})
					logrus.WithFields(logrus.Fields{
						"volume":   blockDevice["volume_id"],
						"instance": instance.TerraformId(),
					}).Debug("Creating aws_ebs_volume from aws_instance.ebs_block_device")
					data := map[string]interface{}{
						"availability_zone": func() interface{} {
							if val, ok := instance.Attrs.Get("availability_zone"); ok {
								return val
							}
							return nil
						}(),
						"encrypted":            blockDevice["encrypted"],
						"id":                   blockDevice["volume_id"],
						"iops":                 blockDevice["iops"],
						"kms_key_id":           blockDevice["kms_key_id"],
						"size":                 blockDevice["volume_size"],
						"type":                 blockDevice["volume_type"],
						"multi_attach_enabled": false,
						"tags": func() interface{} {
							if val, ok := instance.Attrs.Get("volume_tags"); ok {
								return val
							}
							return nil
						}(),
					}
					ctyVal, err := a.resourceFactory.CreateResource(data, "aws_ebs_volume")
					if err != nil {
						return err
					}
					ebsVolume := aws.AwsEbsVolume{
						AvailabilityZone: func() *string {
							if val, ok := instance.Attrs.Get("availability_zone"); ok {
								return awssdk.String(val.(string))
							}
							return nil
						}(),
						Encrypted: func() *bool {
							if val, ok := blockDevice["encrypted"]; ok {
								return awssdk.Bool(val.(bool))
							}
							return nil
						}(),
						Id: blockDevice["volume_id"].(string),
						Iops: func() *int {
							if val, ok := blockDevice["iops"]; ok {
								return awssdk.Int(int(val.(float64)))
							}
							return nil
						}(),
						KmsKeyId: func() *string {
							if val, ok := blockDevice["kms_key_id"]; ok {
								return awssdk.String(val.(string))
							}
							return nil
						}(),
						Size: func() *int {
							if val, ok := blockDevice["volume_size"]; ok {
								return awssdk.Int(int(val.(float64)))
							}
							return nil
						}(),
						Type: func() *string {
							if val, ok := blockDevice["volume_type"]; ok {
								return awssdk.String(val.(string))
							}
							return nil
						}(),
						MultiAttachEnabled: awssdk.Bool(false),
						Tags: func() map[string]string {
							if val, ok := instance.Attrs.Get("volume_tags"); ok {
								return val.(map[string]string)
							}
							return nil
						}(),
						CtyVal: ctyVal,
					}
					newStateResources = append(newStateResources, &ebsVolume)
				}
				instance.Attrs.SafeDelete([]string{"ebs_block_device"})
			}
			newStateResources = append(newStateResources, instance)
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

	if resource.IsRefactoredResource("aws_instance") {
		newRemoteResources := make([]resource.Resource, 0)
		for _, remoteRes := range *remoteResources {
			if remoteRes.TerraformType() != aws.AwsInstanceResourceType {
				newRemoteResources = append(newRemoteResources, remoteRes)
				continue
			}
			instance, _ := remoteRes.(*resource.AbstractResource)
			instance.Attrs.SafeDelete([]string{"root_block_device"})
			instance.Attrs.SafeDelete([]string{"ebs_block_device"})
			newRemoteResources = append(newRemoteResources, instance)
		}
		*remoteResources = newRemoteResources
	}

	*resourcesFromState = newStateResources

	return nil
}

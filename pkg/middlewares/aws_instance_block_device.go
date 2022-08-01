package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// Remove root_block_device from aws_instance resources and create dedicated aws_ebs_volume resources
type AwsInstanceBlockDeviceResourceMapper struct {
	resourceFactory resource.ResourceFactory
}

func NewAwsInstanceBlockDeviceResourceMapper(resourceFactory resource.ResourceFactory) AwsInstanceBlockDeviceResourceMapper {
	return AwsInstanceBlockDeviceResourceMapper{
		resourceFactory: resourceFactory,
	}
}

func (a AwsInstanceBlockDeviceResourceMapper) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newStateResources := make([]*resource.Resource, 0)
	for _, stateRes := range *resourcesFromState {

		// Ignore all resources other than aws_instance
		if stateRes.ResourceType() != aws.AwsInstanceResourceType {
			newStateResources = append(newStateResources, stateRes)
			continue
		}

		if rootBlockDevice, exist := stateRes.Attrs.Get("root_block_device"); exist {
			for _, rootBlock := range rootBlockDevice.([]interface{}) {
				rootBlock := rootBlock.(map[string]interface{})
				logrus.WithFields(logrus.Fields{
					"volume":   rootBlock["volume_id"],
					"instance": stateRes.ResourceId(),
				}).Debug("Creating aws_ebs_volume from aws_instance.root_block_device")
				data := map[string]interface{}{
					"availability_zone":    (*stateRes.Attrs)["availability_zone"],
					"encrypted":            rootBlock["encrypted"],
					"id":                   rootBlock["volume_id"],
					"iops":                 rootBlock["iops"],
					"kms_key_id":           rootBlock["kms_key_id"],
					"size":                 rootBlock["volume_size"],
					"type":                 rootBlock["volume_type"],
					"multi_attach_enabled": false,
					"tags":                 a.volumeTags(stateRes, rootBlock),
				}
				if throughput, exist := rootBlock["throughput"]; exist {
					data["throughput"] = throughput
				}
				newRes := a.resourceFactory.CreateAbstractResource("aws_ebs_volume", rootBlock["volume_id"].(string), data)
				newStateResources = append(newStateResources, newRes)
			}
			stateRes.Attrs.SafeDelete([]string{"root_block_device"})
			stateRes.Attrs.SafeDelete([]string{"volume_tags"})
		}

		if ebsBlockDevice, exist := stateRes.Attrs.Get("ebs_block_device"); exist {
			for _, blockDevice := range ebsBlockDevice.([]interface{}) {
				blockDevice := blockDevice.(map[string]interface{})
				if a.hasBlockDevice(blockDevice, resourcesFromState) {
					continue
				}
				logrus.WithFields(logrus.Fields{
					"volume":   blockDevice["volume_id"],
					"instance": stateRes.ResourceId(),
				}).Debug("Creating aws_ebs_volume from aws_instance.ebs_block_device")
				data := map[string]interface{}{
					"availability_zone":    (*stateRes.Attrs)["availability_zone"],
					"encrypted":            blockDevice["encrypted"],
					"id":                   blockDevice["volume_id"],
					"iops":                 blockDevice["iops"],
					"kms_key_id":           blockDevice["kms_key_id"],
					"size":                 blockDevice["volume_size"],
					"type":                 blockDevice["volume_type"],
					"multi_attach_enabled": false,
					"tags":                 a.volumeTags(stateRes, blockDevice),
				}
				if throughput, exist := blockDevice["throughput"]; exist {
					data["throughput"] = throughput
				}
				newRes := a.resourceFactory.CreateAbstractResource("aws_ebs_volume", blockDevice["volume_id"].(string), data)
				newStateResources = append(newStateResources, newRes)
			}
			stateRes.Attrs.SafeDelete([]string{"ebs_block_device"})
			stateRes.Attrs.SafeDelete([]string{"volume_tags"})
		}
		newStateResources = append(newStateResources, stateRes)
	}

	newRemoteResources := make([]*resource.Resource, 0)
	for _, remoteRes := range *remoteResources {
		if remoteRes.ResourceType() != aws.AwsInstanceResourceType {
			newRemoteResources = append(newRemoteResources, remoteRes)
			continue
		}
		remoteRes.Attrs.SafeDelete([]string{"root_block_device"})
		remoteRes.Attrs.SafeDelete([]string{"ebs_block_device"})
		remoteRes.Attrs.SafeDelete([]string{"volume_tags"})
		newRemoteResources = append(newRemoteResources, remoteRes)
	}

	*resourcesFromState = newStateResources
	*remoteResources = newRemoteResources

	return nil
}

func (a AwsInstanceBlockDeviceResourceMapper) volumeTags(instance *resource.Resource, blockDevice map[string]interface{}) interface{} {
	if tags, exist := instance.Attrs.Get("volume_tags"); exist {
		return tags
	}
	return blockDevice["tags"]
}

func (a AwsInstanceBlockDeviceResourceMapper) hasBlockDevice(blockDevice map[string]interface{}, resourcesFromState *[]*resource.Resource) bool {
	for _, stateRes := range *resourcesFromState {
		if stateRes.ResourceType() == aws.AwsEbsVolumeResourceType &&
			stateRes.ResourceId() == blockDevice["volume_id"] {
			return true
		}
	}
	return false
}

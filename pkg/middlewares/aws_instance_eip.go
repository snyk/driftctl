package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

type AwsInstanceEIP struct{}

func (a AwsInstanceEIP) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than aws_instance
		if remoteResource.TerraformType() != aws.AwsInstanceResourceType {
			continue
		}

		instance, _ := remoteResource.(*aws.AwsInstance)

		if a.hasEIP(instance, resourcesFromState) {
			logrus.WithFields(logrus.Fields{
				"instance": instance.TerraformId(),
			}).Debug("Ignore instance public ip and dns as it has an eip attached")
			a.ignorePublicIpAndDns(instance, remoteResources, resourcesFromState)
		}
	}

	return nil
}

func (a AwsInstanceEIP) hasEIP(instance *aws.AwsInstance, resources *[]resource.Resource) bool {
	for _, res := range *resources {
		if res.TerraformType() == aws.AwsEipResourceType {
			eip, _ := res.(*aws.AwsEip)
			if *eip.Instance == instance.Id {
				return true
			}
		}
		if res.TerraformType() == aws.AwsEipAssociationResourceType {
			eip, _ := res.(*aws.AwsEipAssociation)
			if *eip.InstanceId == instance.Id {
				return true
			}
		}
	}

	return false
}

func (a AwsInstanceEIP) ignorePublicIpAndDns(instance *aws.AwsInstance, resourcesSet ...*[]resource.Resource)  {
	for _, resources := range resourcesSet {
		for _, res := range *resources {
			if res.TerraformType() == instance.TerraformType() &&
				res.TerraformId() == instance.TerraformId() {
				instance, _ := res.(*aws.AwsInstance)
				instance.PublicDns = nil
				instance.PublicIp = nil
			}
		}
	}
}

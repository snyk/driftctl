package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

type AwsInstanceEIP struct{}

func (a AwsInstanceEIP) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than aws_instance
		if remoteResource.TerraformType() != aws.AwsInstanceResourceType {
			continue
		}

		if a.hasEIP(remoteResource, resourcesFromState) {
			logrus.WithFields(logrus.Fields{
				"instance": remoteResource.TerraformId(),
			}).Debug("Ignore instance public ip and dns as it has an eip attached")
			a.ignorePublicIpAndDns(remoteResource, remoteResources, resourcesFromState)
		}
	}

	return nil
}

func (a AwsInstanceEIP) hasEIP(instance *resource.Resource, resources *[]*resource.Resource) bool {
	for _, res := range *resources {
		if res.TerraformType() == aws.AwsEipResourceType {
			if (*res.Attrs)["instance"] == instance.TerraformId() {
				return true
			}
		}
		if res.TerraformType() == aws.AwsEipAssociationResourceType {
			if (*res.Attrs)["instance_id"] == instance.TerraformId() {
				return true
			}
		}
	}

	return false
}

func (a AwsInstanceEIP) ignorePublicIpAndDns(instance *resource.Resource, resourcesSet ...*[]*resource.Resource) {
	for _, resources := range resourcesSet {
		for _, res := range *resources {
			if res.TerraformType() == instance.TerraformType() &&
				res.TerraformId() == instance.TerraformId() {
				res.Attrs.SafeDelete([]string{"public_dns"})
				res.Attrs.SafeDelete([]string{"public_ip"})
			}
		}
	}
}

package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

const defaultAwsSecurityGroupName = "default"

// When scanning a brand new AWS account, some users may see irrelevant results about default AWS security group.
// We ignore these resources by default when strict mode is disabled.
type AwsSecurityGroupDefaults struct{}

func NewAwsSecurityGroupDefaults() AwsSecurityGroupDefaults {
	return AwsSecurityGroupDefaults{}
}

func (m AwsSecurityGroupDefaults) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	newRemoteResources := make([]resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than security group
		if remoteResource.TerraformType() != aws.AwsSecurityGroupResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		existInState := false
		for _, stateResource := range *resourcesFromState {
			if resource.IsSameResource(remoteResource, stateResource) {
				existInState = true
				break
			}
		}

		if existInState || *remoteResource.(*aws.AwsSecurityGroup).Name != defaultAwsSecurityGroupName {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		logrus.WithFields(logrus.Fields{
			"id":   remoteResource.TerraformId(),
			"type": remoteResource.TerraformType(),
		}).Debug("Ignoring default aws security group as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

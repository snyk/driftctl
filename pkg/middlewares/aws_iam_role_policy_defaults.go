package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// Default subnet should not be shown as unmanaged as they are present by default
// This middleware ignores default subnet from unmanaged resources if they are not managed by IaC
type AwsDefaultIamRolePolicy struct{}

func NewAwsDefaultIamRolePolicy() AwsDefaultIamRolePolicy {
	return AwsDefaultIamRolePolicy{}
}

func (m AwsDefaultIamRolePolicy) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {

	newRemoteResources := make([]resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		existInState := false

		// Ignore all resources other than default Subnet
		if remoteResource.TerraformType() != aws.AwsIamRolePolicyResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		for _, stateResource := range *resourcesFromState {
			if resource.IsSameResource(remoteResource, stateResource) {
				existInState = true
				break
			}
		}

		if existInState {
			newRemoteResources = append(newRemoteResources, remoteResource)
		} else {
			logrus.WithFields(logrus.Fields{
				"id":   remoteResource.TerraformId(),
				"type": remoteResource.TerraformType(),
			}).Debug("Ignoring default IAM policies as it is not managed by IaC")
		}

	}

	*remoteResources = newRemoteResources

	return nil
}

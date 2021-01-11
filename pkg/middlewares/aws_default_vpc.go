package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// Default VPC should not be shown as unmanaged as they are present by default
// This middleware ignores default VPC from unmanaged resources if they are not managed by IaC
type AwsDefaultVPC struct{}

func NewAwsDefaultVPC() AwsDefaultVPC {
	return AwsDefaultVPC{}
}

func (m AwsDefaultVPC) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {

	newRemoteResources := make([]resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		existInState := false

		// Ignore all resources other than default VPC
		if remoteResource.TerraformType() != aws.AwsDefaultVpcResourceType {
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
		}

		if !existInState {
			logrus.WithFields(logrus.Fields{
				"id":   remoteResource.TerraformId(),
				"type": remoteResource.TerraformType(),
			}).Debug("Ignoring default VPC as it is not managed by IaC")
		}

	}

	*remoteResources = newRemoteResources

	return nil
}

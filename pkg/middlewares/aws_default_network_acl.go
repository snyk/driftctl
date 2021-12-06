package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// Default network ACL should not be shown as unmanaged as they are present by default
// This middleware ignores default network ACL from unmanaged resources if they are not managed by IaC
type AwsDefaultNetworkACL struct{}

func NewAwsDefaultNetworkACL() AwsDefaultNetworkACL {
	return AwsDefaultNetworkACL{}
}

func (m AwsDefaultNetworkACL) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newRemoteResources := make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than network ACLs
		if remoteResource.ResourceType() != aws.AwsDefaultNetworkACLResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Check if resource is managed by IaC
		existInState := false
		for _, stateResource := range *resourcesFromState {
			if remoteResource.Equal(stateResource) {
				existInState = true
				break
			}
		}

		// Include resource if it's managed in IaC
		if existInState {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Else, resource is not added to newRemoteResources slice so it will be ignored
		logrus.WithFields(logrus.Fields{
			"id":   remoteResource.ResourceId(),
			"type": remoteResource.ResourceType(),
		}).Debug("Ignoring default network ACL as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

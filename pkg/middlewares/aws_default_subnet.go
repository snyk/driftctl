package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// Default subnet should not be shown as unmanaged as they are present by default
// This middleware ignores default subnet from unmanaged resources if they are not managed by IaC
type AwsDefaultSubnet struct{}

func NewAwsDefaultSubnet() AwsDefaultSubnet {
	return AwsDefaultSubnet{}
}

func (m AwsDefaultSubnet) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newRemoteResources := make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		existInState := false

		// Ignore all resources other than default Subnet
		if remoteResource.ResourceType() != aws.AwsDefaultSubnetResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		for _, stateResource := range *resourcesFromState {
			if remoteResource.Equal(stateResource) {
				existInState = true
				break
			}
		}

		if existInState {
			newRemoteResources = append(newRemoteResources, remoteResource)
		}

		if !existInState {
			logrus.WithFields(logrus.Fields{
				"id":   remoteResource.ResourceId(),
				"type": remoteResource.ResourceType(),
			}).Debug("Ignoring default Subnet as it is not managed by IaC")
		}

	}

	*remoteResources = newRemoteResources

	return nil
}

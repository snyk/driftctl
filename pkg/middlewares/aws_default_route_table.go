package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

// Default route table should not be shown as unmanaged as they are present by default
// This middleware ignores default route table from unmanaged resources if they are not managed by IaC
type AwsDefaultRouteTable struct{}

func NewAwsDefaultRouteTable() AwsDefaultRouteTable {
	return AwsDefaultRouteTable{}
}

func (m AwsDefaultRouteTable) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newRemoteResources := make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		existInState := false

		// Ignore all resources other than default RouteTable
		if remoteResource.ResourceType() != aws.AwsDefaultRouteTableResourceType {
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
			}).Debug("Ignoring default route table as it is not managed by IaC")
		}

	}

	*remoteResources = newRemoteResources

	return nil
}

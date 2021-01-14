package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// Default route table should not be shown as unmanaged as they are present by default
// This middleware ignores default route table from unmanaged resources if they are not managed by IaC
type AwsDefaultRouteTable struct{}

func NewAwsDefaultRouteTable() AwsDefaultRouteTable {
	return AwsDefaultRouteTable{}
}

func (m AwsDefaultRouteTable) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {

	newRemoteResources := make([]resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		existInState := false

		// Ignore all resources other than default RouteTable
		if remoteResource.TerraformType() != aws.AwsDefaultRouteTableResourceType {
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
			}).Debug("Ignoring default route table as it is not managed by IaC")
		}

	}

	*remoteResources = newRemoteResources

	return nil
}

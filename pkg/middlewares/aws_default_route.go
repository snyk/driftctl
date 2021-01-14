package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// Default routes should not be shown as unmanaged as they are present by default
// This middleware ignores default routes from unmanaged resources if they are not managed by IaC
type AwsDefaultRoute struct{}

func NewAwsDefaultRoute() AwsDefaultRoute {
	return AwsDefaultRoute{}
}

func (m AwsDefaultRoute) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {

	newRemoteResources := make([]resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than routes
		if remoteResource.TerraformType() != aws.AwsRouteResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		route, _ := remoteResource.(*aws.AwsRoute)
		// Ignore all non-default routes, check if route is coming from table creation
		if route.Origin != nil && *route.Origin != "CreateRouteTable" {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Check if route is managed by IaC
		existInState := false
		for _, stateResource := range *resourcesFromState {
			if resource.IsSameResource(remoteResource, stateResource) {
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
			"route": route.String(),
			"id":    route.TerraformId(),
			"type":  route.TerraformType(),
		}).Debug("Ignoring default route as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

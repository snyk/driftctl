package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// Each region has a default vpc which has an internet gateway attached and thus the route table of this
// same vpc has a default route (0.0.0.0/0) that should not be seen as unmanaged if not managed by IaC
// This middleware ignores the above route from unmanaged resources if not managed by IaC
type AwsDefaultInternetGatewayRoute struct{}

func NewAwsDefaultInternetGatewayRoute() AwsDefaultInternetGatewayRoute {
	return AwsDefaultInternetGatewayRoute{}
}

func (m AwsDefaultInternetGatewayRoute) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	newRemoteResources := make([]resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than routes
		if remoteResource.TerraformType() != aws.AwsRouteResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		route, _ := remoteResource.(*resource.AbstractResource)
		// Ignore all routes except the one that came from the default internet gateway
		if !isDefaultInternetGatewayRoute(route, remoteResources) {
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
			// "route": route.String(), TODO
			"id":   route.TerraformId(),
			"type": route.TerraformType(),
		}).Debug("Ignoring default internet gateway route as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

// Return true if the route's target is the default internet gateway (e.g. attached to the default vpc)
func isDefaultInternetGatewayRoute(route *resource.AbstractResource, remoteResources *[]resource.Resource) bool {
	for _, remoteResource := range *remoteResources {
		if remoteResource.TerraformType() == aws.AwsInternetGatewayResourceType &&
			isDefaultInternetGateway(remoteResource.(*aws.AwsInternetGateway), remoteResources) {
			gtwId, gtwIdExist := route.Attrs.Get("gateway_id")
			destCIDRBlock, destCIDRBlockExist := route.Attrs.Get("destination_cidr_block")
			return gtwIdExist && destCIDRBlockExist && gtwId == remoteResource.TerraformId() && destCIDRBlock == "0.0.0.0/0"
		}
	}
	return false
}

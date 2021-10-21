package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// Each API Gateway rest API has by design all the gateway responses available to edit in the console
// which result in useless noises (e.g. lots of unmanaged resources) by driftctl.
// This middleware ignores all console responses if not managed by IAC.
type AwsConsoleApiGatewayGatewayResponse struct{}

func NewAwsConsoleApiGatewayGatewayResponse() AwsConsoleApiGatewayGatewayResponse {
	return AwsConsoleApiGatewayGatewayResponse{}
}

func (m AwsConsoleApiGatewayGatewayResponse) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newRemoteResources := make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than gateway responses
		if remoteResource.ResourceType() != aws.AwsApiGatewayGatewayResponseResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Check if gateway response is managed by IaC
		existInState := false
		for _, stateResource := range *resourcesFromState {
			if remoteResource.Equal(stateResource) {
				existInState = true
				break
			}
		}

		// Include resource if it's managed by IaC
		if existInState {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Else, resource is not added to newRemoteResources slice so it will be ignored
		logrus.WithFields(logrus.Fields{
			"id":   remoteResource.ResourceId(),
			"type": remoteResource.ResourceType(),
		}).Debug("Ignoring default api gateway response as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

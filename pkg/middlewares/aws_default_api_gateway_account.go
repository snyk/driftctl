package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsDefaultApiGatewayAccount is a middleware that ignores the default API Gateway account resource in the current region.
type AwsDefaultApiGatewayAccount struct{}

func NewAwsDefaultApiGatewayAccount() AwsDefaultApiGatewayAccount {
	return AwsDefaultApiGatewayAccount{}
}

func (m AwsDefaultApiGatewayAccount) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newRemoteResources := make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than API gateway account
		if remoteResource.ResourceType() != aws.AwsApiGatewayAccountResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Check if account is managed by IaC
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

		// Else, resource is not added to newRemoteResources slice, so it will be ignored
		logrus.WithFields(logrus.Fields{
			"id":   remoteResource.ResourceId(),
			"type": remoteResource.ResourceType(),
		}).Debug("Ignoring default API gateway account as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

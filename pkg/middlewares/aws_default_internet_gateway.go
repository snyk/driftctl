package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

// Each default vpc has an internet gateway attached that should not be seen as unmanaged if not managed by IaC
// This middleware ignores default internet gateway from unmanaged resources if not managed by IaC
type AwsDefaultInternetGateway struct{}

func NewAwsDefaultInternetGateway() AwsDefaultInternetGateway {
	return AwsDefaultInternetGateway{}
}

func (m AwsDefaultInternetGateway) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newRemoteResources := make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than internet gateways
		if remoteResource.ResourceType() != aws.AwsInternetGatewayResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Ignore all non-default internet gateways
		if !isDefaultInternetGateway(remoteResource, remoteResources) {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Check if internet gateway is managed by IaC
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
		}).Debug("Ignoring default internet gateway as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

// Return true if the internet gateway is the default one (e.g. attached to the default vpc)
func isDefaultInternetGateway(internetGateway *resource.Resource, remoteResources *[]*resource.Resource) bool {
	for _, remoteResource := range *remoteResources {
		if remoteResource.ResourceType() == aws.AwsDefaultVpcResourceType {
			vpcId, exist := internetGateway.Attrs.Get("vpc_id")
			return exist && vpcId == remoteResource.ResourceId()
		}
	}
	return false
}

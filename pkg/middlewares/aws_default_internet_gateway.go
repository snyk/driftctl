package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// Each default vpc has an internet gateway attached that should not be seen as unmanaged if not managed by IaC
// This middleware ignores default internet gateway from unmanaged resources if not managed by IaC
type AwsDefaultInternetGateway struct{}

func NewAwsDefaultInternetGateway() AwsDefaultInternetGateway {
	return AwsDefaultInternetGateway{}
}

func (m AwsDefaultInternetGateway) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	newRemoteResources := make([]resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than internet gateways
		if remoteResource.TerraformType() != aws.AwsInternetGatewayResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		internetGateway, _ := remoteResource.(*aws.AwsInternetGateway)
		// Ignore all non-default internet gateways
		if !isDefaultInternetGateway(internetGateway, remoteResources) {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Check if internet gateway is managed by IaC
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
			"id":   internetGateway.TerraformId(),
			"type": internetGateway.TerraformType(),
		}).Debug("Ignoring default internet gateway as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

// Return true if the internet gateway is the default one (e.g. attached to the default vpc)
func isDefaultInternetGateway(internetGateway *aws.AwsInternetGateway, remoteResources *[]resource.Resource) bool {
	for _, remoteResource := range *remoteResources {
		if remoteResource.TerraformType() == aws.AwsDefaultVpcResourceType {
			return *internetGateway.VpcId == remoteResource.TerraformId()
		}
	}
	return false
}

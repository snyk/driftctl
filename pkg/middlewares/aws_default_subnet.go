package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
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
		if remoteResource.TerraformType() != aws.AwsDefaultSubnetResourceType {
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
				"id":   remoteResource.TerraformId(),
				"type": remoteResource.TerraformType(),
			}).Debug("Ignoring default Subnet as it is not managed by IaC")
		}

	}

	*remoteResources = newRemoteResources

	return nil
}

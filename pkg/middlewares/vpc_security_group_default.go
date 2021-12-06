package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// Remove default security group from remote resources
type VPCDefaultSecurityGroupSanitizer struct{}

func NewVPCDefaultSecurityGroupSanitizer() VPCDefaultSecurityGroupSanitizer {
	return VPCDefaultSecurityGroupSanitizer{}
}

func (m VPCDefaultSecurityGroupSanitizer) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newRemoteResources := make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		existInState := false

		// Ignore all resources other than default security group
		if remoteResource.ResourceType() != aws.AwsDefaultSecurityGroupResourceType {
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
			}).Debug("Ignoring default unmanaged security group")
		}
	}

	*remoteResources = newRemoteResources

	return nil
}

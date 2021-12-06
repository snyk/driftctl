package middlewares

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
)

// Some service accounts are created by default when activating APIs, this middleware will filter them unless they are managed.
type GoogleDefaultIAMMember struct{}

func NewGoogleDefaultIAMMember() *GoogleDefaultIAMMember {
	return &GoogleDefaultIAMMember{}
}

func (m *GoogleDefaultIAMMember) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newRemoteResources := make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than BucketIamBinding
		if remoteResource.ResourceType() != google.GoogleProjectIamMemberResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Ignore all non service account member
		if memberName := remoteResource.Attrs.GetString("member"); memberName != nil && !strings.HasPrefix(*memberName, "serviceAccount:") {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Ignore all service accounts that have project host
		memberName := *remoteResource.Attrs.GetString("member")
		projectName := *remoteResource.Attrs.GetString("project")

		if strings.HasSuffix(memberName, fmt.Sprintf("%s.iam.gserviceaccount.com", projectName)) {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Check if member is managed by IaC
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

		// Else, resource is not added to newRemoteResources slice, so it will be ignored
		logrus.WithFields(logrus.Fields{
			"id":   remoteResource.ResourceId(),
			"type": remoteResource.ResourceType(),
		}).Debug("Ignoring default service account membership as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

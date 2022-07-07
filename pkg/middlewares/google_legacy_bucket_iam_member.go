package middlewares

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
)

// Creating buckets add legacy role bindings, this middleware will filter them unless they are managed.
type GoogleLegacyBucketIAMMember struct{}

func NewGoogleLegacyBucketIAMMember() *GoogleLegacyBucketIAMMember {
	return &GoogleLegacyBucketIAMMember{}
}

func (m *GoogleLegacyBucketIAMMember) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newRemoteResources := make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than BucketIamBinding
		if remoteResource.ResourceType() != google.GoogleStorageBucketIamMemberResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Ignore all non-legacy member
		if roleName := remoteResource.Attrs.GetString("role"); roleName != nil && !strings.Contains(*roleName, "legacy") {
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

		// Include resource if it's managed in IaC
		if existInState {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Else, resource is not added to newRemoteResources slice, so it will be ignored
		logrus.WithFields(logrus.Fields{
			"id":   remoteResource.ResourceId(),
			"type": remoteResource.ResourceType(),
		}).Debug("Ignoring legacy bucket member as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

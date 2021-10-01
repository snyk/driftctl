package middlewares

import (
	"strings"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
	"github.com/sirupsen/logrus"
)

// Creating buckets add legacy role bindings, this middleware will filter them unless they are managed.
type GoogleLegacyBucketIAMBindings struct{}

func NewGoogleLegacyBucketIAMBindings() GoogleLegacyBucketIAMBindings {
	return GoogleLegacyBucketIAMBindings{}
}

func (m GoogleLegacyBucketIAMBindings) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newRemoteResources := make([]*resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than BucketIamBinding
		if remoteResource.ResourceType() != google.GoogleStorageBucketIamBindingResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Ignore all non-legacy bindings
		if roleName := remoteResource.Attrs.GetString("role"); roleName != nil && !strings.Contains(*roleName, "legacy") {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		// Check if bindings is managed by IaC
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
		}).Debug("Ignoring legacy bucket bindings as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

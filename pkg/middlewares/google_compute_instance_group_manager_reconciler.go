package middlewares

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
)

type GoogleComputeInstanceGroupManagerReconciler struct{}

// NewGoogleComputeInstanceGroupManagerReconciler imports remote instance groups when they're managed by a managed instance group manager.
// Creating a "google_compute_instance_group_manager" resource via Terraform leads to having several unmanaged instance groups.
// This middleware adds remote instance groups to the state by matching them with managed instance group managers.
func NewGoogleComputeInstanceGroupManagerReconciler() *GoogleComputeInstanceGroupManagerReconciler {
	return &GoogleComputeInstanceGroupManagerReconciler{}
}

func (a GoogleComputeInstanceGroupManagerReconciler) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	var newStateResources []*resource.Resource

	instanceGroups := make([]*resource.Resource, 0)
	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than google_compute_instance_group
		if remoteResource.ResourceType() != google.GoogleComputeInstanceGroupResourceType {
			continue
		}
		instanceGroups = append(instanceGroups, remoteResource)
	}

	for _, stateResource := range *resourcesFromState {
		newStateResources = append(newStateResources, stateResource)

		// Ignore all resources other than google_compute_instance_group_manager
		if stateResource.ResourceType() != google.GoogleComputeInstanceGroupManagerResourceType {
			continue
		}

		name := stateResource.Attributes().GetString("name")

		for _, group := range instanceGroups {
			// Import instance group in the state
			if n := group.Attributes().GetString("name"); n != nil && *n == *name {
				newStateResources = append(newStateResources, group)
			}
		}
	}

	*resourcesFromState = newStateResources

	return nil
}

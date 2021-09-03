package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// AwsRDSClusterInstanceExpander search for cluster instances from state to import corresponding remote db instances.
// RDS cluster instance does not represent an actual AWS resource, so shouldn't be used for comparison.
type AwsRDSClusterInstanceExpander struct {
	resourceFactory resource.ResourceFactory
}

func NewRDSClusterInstanceExpander(resourceFactory resource.ResourceFactory) AwsRDSClusterInstanceExpander {
	return AwsRDSClusterInstanceExpander{
		resourceFactory: resourceFactory,
	}
}

func (m AwsRDSClusterInstanceExpander) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newRemoteResources := make([]*resource.Resource, 0)
	newResourcesFromState := make([]*resource.Resource, 0)

	dbInstances := make([]*resource.Resource, 0)
	for _, remoteRes := range *remoteResources {
		if remoteRes.ResourceType() != aws.AwsDbInstanceResourceType {
			newRemoteResources = append(newRemoteResources, remoteRes)
			continue
		}
		dbInstances = append(dbInstances, remoteRes)
	}

	for _, res := range *resourcesFromState {
		// Ignore all resources other than rds_cluster_instance
		if res.ResourceType() != aws.AwsRDSClusterInstanceResourceType {
			newResourcesFromState = append(newResourcesFromState, res)
			continue
		}

		var found bool
		for _, remoteRes := range dbInstances {
			// If the db instance's id matches the rds cluster instance's id, import it in the state
			if remoteRes.ResourceId() == res.ResourceId() {
				newRemoteResources = append(newRemoteResources, remoteRes)
				newResourcesFromState = append(newResourcesFromState, remoteRes)
				found = true
			}
		}

		if !found {
			newResourcesFromState = append(newResourcesFromState, res)
		}
	}
	*resourcesFromState = newResourcesFromState
	*remoteResources = newRemoteResources
	return nil
}

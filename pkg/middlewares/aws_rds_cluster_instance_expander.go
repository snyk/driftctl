package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
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
	newResourcesFromState := make([]*resource.Resource, 0)

	dbInstances := make([]*resource.Resource, 0)
	for _, remoteRes := range *remoteResources {
		if remoteRes.ResourceType() != aws.AwsDbInstanceResourceType {
			continue
		}
		dbInstances = append(dbInstances, remoteRes)
	}

	for _, stateRes := range *resourcesFromState {
		// Ignore all resources other than rds_cluster_instance
		if stateRes.ResourceType() != aws.AwsRDSClusterInstanceResourceType {
			newResourcesFromState = append(newResourcesFromState, stateRes)
			continue
		}

		var found bool
		for _, remoteRes := range dbInstances {
			// If the db instance's id matches the rds cluster instance's id, import it in the state
			if remoteRes.ResourceId() == stateRes.ResourceId() {
				found = true
				newDbInstance := m.resourceFactory.CreateAbstractResource(aws.AwsDbInstanceResourceType, remoteRes.ResourceId(), *remoteRes.Attributes())
				newResourcesFromState = append(newResourcesFromState, newDbInstance)
				logrus.WithFields(logrus.Fields{
					"id": newDbInstance.ResourceId(),
				}).Debug("Created new db instance from RDS cluster instance")
				break
			}
		}

		// If we don't manage to find a db instance corresponding to this RDS cluster instance, simply add it back to the state.
		if !found {
			newResourcesFromState = append(newResourcesFromState, stateRes)
		}
	}
	*resourcesFromState = newResourcesFromState
	return nil
}

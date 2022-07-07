package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type RDSClusterEnumerator struct {
	repository repository.RDSRepository
	factory    resource.ResourceFactory
}

func NewRDSClusterEnumerator(repository repository.RDSRepository, factory resource.ResourceFactory) *RDSClusterEnumerator {
	return &RDSClusterEnumerator{
		repository,
		factory,
	}
}

func (e *RDSClusterEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsRDSClusterResourceType
}

func (e *RDSClusterEnumerator) Enumerate() ([]*resource.Resource, error) {
	clusters, err := e.repository.ListAllDBClusters()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(clusters))

	for _, cluster := range clusters {
		var databaseName string

		if v := cluster.DatabaseName; v != nil {
			databaseName = *cluster.DatabaseName
		}

		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*cluster.DBClusterIdentifier,
				map[string]interface{}{
					"cluster_identifier": *cluster.DBClusterIdentifier,
					"database_name":      databaseName,
				},
			),
		)
	}

	return results, nil
}

package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
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

	results := make([]*resource.Resource, len(clusters))

	for _, cluster := range clusters {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*cluster.DBClusterIdentifier,
				map[string]interface{}{
					"cluster_identifier": *cluster.DBClusterIdentifier,
				},
			),
		)
	}

	return results, nil
}

package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type ElastiCacheClusterEnumerator struct {
	repository repository.ElastiCacheRepository
	factory    resource.ResourceFactory
}

func NewElastiCacheClusterEnumerator(repo repository.ElastiCacheRepository, factory resource.ResourceFactory) *ElastiCacheClusterEnumerator {
	return &ElastiCacheClusterEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ElastiCacheClusterEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsElastiCacheClusterResourceType
}

func (e *ElastiCacheClusterEnumerator) Enumerate() ([]*resource.Resource, error) {
	clusters, err := e.repository.ListAllCacheClusters()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(clusters))

	for _, cluster := range clusters {
		c := cluster
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*c.CacheClusterId,
				map[string]interface{}{},
			),
		)
	}
	return results, err
}

package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	"github.com/snyk/driftctl/pkg/remote/cache"
)

type ElastiCacheRepository interface {
	ListAllCacheClusters() ([]*elasticache.CacheCluster, error)
}

type elasticacheRepository struct {
	client elasticacheiface.ElastiCacheAPI
	cache  cache.Cache
}

func NewElastiCacheRepository(session *session.Session, c cache.Cache) *elasticacheRepository {
	return &elasticacheRepository{
		elasticache.New(session),
		c,
	}
}

func (r *elasticacheRepository) ListAllCacheClusters() ([]*elasticache.CacheCluster, error) {
	if v := r.cache.Get("elasticacheListAllCacheClusters"); v != nil {
		return v.([]*elasticache.CacheCluster), nil
	}

	var clusters []*elasticache.CacheCluster
	input := elasticache.DescribeCacheClustersInput{}
	err := r.client.DescribeCacheClustersPages(&input,
		func(resp *elasticache.DescribeCacheClustersOutput, lastPage bool) bool {
			clusters = append(clusters, resp.CacheClusters...)
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}

	r.cache.Put("elasticacheListAllCacheClusters", clusters)
	return clusters, nil
}

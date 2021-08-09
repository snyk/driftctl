package repository

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/eks/eksiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type EKSRepository interface {
	ListAllClusters() ([]*string, error)
	DescribeCluster(string) (*eks.Cluster, error)
}

type eksRepository struct {
	client eksiface.EKSAPI
	cache  cache.Cache
}

func NewEKSRepository(session *session.Session, c cache.Cache) *eksRepository {
	return &eksRepository{
		eks.New(session),
		c,
	}
}

func (r *eksRepository) ListAllClusters() ([]*string, error) {
	if v := r.cache.Get("eksListAllClusters"); v != nil {
		return v.([]*string), nil
	}

	input := &eks.ListClustersInput{}
	clusters, err := r.client.ListClusters(input)
	if err != nil {
		return nil, err
	}
	r.cache.Put("eksListAllClusters", clusters.Clusters)
	return clusters.Clusters, err
}

func (r *eksRepository) DescribeCluster(name string) (*eks.Cluster, error) {
	cacheKey := fmt.Sprintf("eksDescribeCluster_%s", name)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.(*eks.Cluster), nil
	}

	input := &eks.DescribeClusterInput{
		Name: &name,
	}
	cluster, err := r.client.DescribeCluster(input)
	if err != nil {
		return nil, err
	}
	r.cache.Put(cacheKey, cluster.Cluster)
	return cluster.Cluster, err
}

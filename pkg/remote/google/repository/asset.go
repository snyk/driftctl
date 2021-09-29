package repository

import (
	"context"
	"fmt"
	"sync"

	asset "cloud.google.com/go/asset/apiv1"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/google/config"
	"google.golang.org/api/iterator"
	assetpb "google.golang.org/genproto/googleapis/cloud/asset/v1"
)

const (
	storageBucketAssetType   = "storage.googleapis.com/Bucket"
	computeFirewallAssetType = "compute.googleapis.com/Firewall"
	computeRouterAssetType   = "compute.googleapis.com/Router"
	computeInstanceAssetType = "compute.googleapis.com/Instance"
	computeNetworkAssetType  = "compute.googleapis.com/Network"
)

type AssetRepository interface {
	SearchAllBuckets() ([]*assetpb.ResourceSearchResult, error)
	SearchAllFirewalls() ([]*assetpb.ResourceSearchResult, error)
	SearchAllRouters() ([]*assetpb.ResourceSearchResult, error)
	SearchAllInstances() ([]*assetpb.ResourceSearchResult, error)
	SearchAllNetworks() ([]*assetpb.ResourceSearchResult, error)
}

type assetRepository struct {
	client *asset.Client
	config config.GCPTerraformConfig
	cache  cache.Cache
	lock   sync.Locker
}

func NewAssetRepository(client *asset.Client, config config.GCPTerraformConfig, c cache.Cache) *assetRepository {
	return &assetRepository{
		client,
		config,
		c,
		&sync.Mutex{},
	}
}

func (s assetRepository) searchAllResources(ty string) ([]*assetpb.ResourceSearchResult, error) {
	req := &assetpb.SearchAllResourcesRequest{
		Scope: fmt.Sprintf("projects/%s", s.config.Project),
		AssetTypes: []string{
			storageBucketAssetType,
			computeFirewallAssetType,
			computeRouterAssetType,
			computeInstanceAssetType,
			computeNetworkAssetType,
		},
	}
	var results []*assetpb.ResourceSearchResult

	s.lock.Lock()
	defer s.lock.Unlock()
	if cachedResults := s.cache.Get("SearchAllResources"); cachedResults != nil {
		results = cachedResults.([]*assetpb.ResourceSearchResult)
	}

	if results == nil {
		it := s.client.SearchAllResources(context.Background(), req)
		for {
			resource, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, err
			}
			results = append(results, resource)
		}
		s.cache.Put("SearchAllResources", results)
	}

	filteredResults := []*assetpb.ResourceSearchResult{}
	for _, result := range results {
		if result.AssetType == ty {
			filteredResults = append(filteredResults, result)
		}
	}

	return filteredResults, nil
}

func (s assetRepository) SearchAllBuckets() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(storageBucketAssetType)
}

func (s assetRepository) SearchAllFirewalls() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(computeFirewallAssetType)
}

func (s assetRepository) SearchAllRouters() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(computeRouterAssetType)
}

func (s assetRepository) SearchAllInstances() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(computeInstanceAssetType)
}

func (s assetRepository) SearchAllNetworks() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(computeNetworkAssetType)
}

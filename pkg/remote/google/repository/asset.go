package repository

import (
	"context"
	"fmt"

	asset "cloud.google.com/go/asset/apiv1"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/google/config"
	"google.golang.org/api/iterator"
	assetpb "google.golang.org/genproto/googleapis/cloud/asset/v1"
)

// https://cloud.google.com/asset-inventory/docs/supported-asset-types#supported_resource_types
const (
	storageBucketAssetType        = "storage.googleapis.com/Bucket"
	computeFirewallAssetType      = "compute.googleapis.com/Firewall"
	computeRouterAssetType        = "compute.googleapis.com/Router"
	computeInstanceAssetType      = "compute.googleapis.com/Instance"
	computeNetworkAssetType       = "compute.googleapis.com/Network"
	dnsManagedZoneAssetType       = "dns.googleapis.com/ManagedZone"
	computeInstanceGroupAssetType = "compute.googleapis.com/InstanceGroup"
	bigqueryDatasetAssetType      = "bigquery.googleapis.com/Dataset"
	bigqueryTableAssetType        = "bigquery.googleapis.com/Table"
	computeAddressAssetType       = "compute.googleapis.com/Address"
)

type AssetRepository interface {
	SearchAllBuckets() ([]*assetpb.ResourceSearchResult, error)
	SearchAllFirewalls() ([]*assetpb.ResourceSearchResult, error)
	SearchAllRouters() ([]*assetpb.ResourceSearchResult, error)
	SearchAllInstances() ([]*assetpb.ResourceSearchResult, error)
	SearchAllNetworks() ([]*assetpb.ResourceSearchResult, error)
	SearchAllDNSManagedZones() ([]*assetpb.ResourceSearchResult, error)
	SearchAllInstanceGroups() ([]*assetpb.ResourceSearchResult, error)
	SearchAllDatasets() ([]*assetpb.ResourceSearchResult, error)
	SearchAllTables() ([]*assetpb.ResourceSearchResult, error)
	SearchAllAddresses() ([]*assetpb.ResourceSearchResult, error)
}

type assetRepository struct {
	client *asset.Client
	config config.GCPTerraformConfig
	cache  cache.Cache
}

func NewAssetRepository(client *asset.Client, config config.GCPTerraformConfig, c cache.Cache) *assetRepository {
	return &assetRepository{
		client,
		config,
		c,
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
			dnsManagedZoneAssetType,
			computeInstanceGroupAssetType,
			bigqueryDatasetAssetType,
			bigqueryTableAssetType,
			computeAddressAssetType,
		},
	}
	var results []*assetpb.ResourceSearchResult

	cacheKey := "SearchAllResources"
	cachedResults := s.cache.GetAndLock(cacheKey)
	defer s.cache.Unlock(cacheKey)
	if cachedResults != nil {
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
		s.cache.Put(cacheKey, results)
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

func (s assetRepository) SearchAllDNSManagedZones() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(dnsManagedZoneAssetType)
}

func (s assetRepository) SearchAllInstanceGroups() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(computeInstanceGroupAssetType)
}

func (s assetRepository) SearchAllDatasets() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(bigqueryDatasetAssetType)
}

func (s assetRepository) SearchAllTables() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(bigqueryTableAssetType)
}

func (s assetRepository) SearchAllAddresses() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(computeAddressAssetType)
}

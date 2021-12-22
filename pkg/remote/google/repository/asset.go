package repository

import (
	"context"
	"errors"
	"fmt"

	asset "cloud.google.com/go/asset/apiv1"
	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/snyk/driftctl/pkg/remote/google/config"
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
	computeSubnetworkAssetType    = "compute.googleapis.com/Subnetwork"
	computeDiskAssetType          = "compute.googleapis.com/Disk"
	computeImageAssetType         = "compute.googleapis.com/Image"
	dnsManagedZoneAssetType       = "dns.googleapis.com/ManagedZone"
	computeInstanceGroupAssetType = "compute.googleapis.com/InstanceGroup"
	bigqueryDatasetAssetType      = "bigquery.googleapis.com/Dataset"
	bigqueryTableAssetType        = "bigquery.googleapis.com/Table"
	computeAddressAssetType       = "compute.googleapis.com/Address"
	computeGlobalAddressAssetType = "compute.googleapis.com/GlobalAddress"
	cloudFunctionsFunction        = "cloudfunctions.googleapis.com/CloudFunction"
	bigtableInstanceAssetType     = "bigtableadmin.googleapis.com/Instance"
	bigtableTableAssetType        = "bigtableadmin.googleapis.com/Table"
	sqlDatabaseInstanceAssetType  = "sqladmin.googleapis.com/Instance"
	healthCheckAssetType          = "compute.googleapis.com/HealthCheck"
	cloudRunServiceAssetType      = "run.googleapis.com/Service"
	nodeGroupAssetType            = "compute.googleapis.com/NodeGroup"
)

type AssetRepository interface {
	SearchAllBuckets() ([]*assetpb.ResourceSearchResult, error)
	SearchAllFirewalls() ([]*assetpb.ResourceSearchResult, error)
	SearchAllRouters() ([]*assetpb.ResourceSearchResult, error)
	SearchAllInstances() ([]*assetpb.ResourceSearchResult, error)
	SearchAllNetworks() ([]*assetpb.ResourceSearchResult, error)
	SearchAllDisks() ([]*assetpb.ResourceSearchResult, error)
	SearchAllImages() ([]*assetpb.ResourceSearchResult, error)
	SearchAllDNSManagedZones() ([]*assetpb.ResourceSearchResult, error)
	SearchAllInstanceGroups() ([]*assetpb.ResourceSearchResult, error)
	SearchAllDatasets() ([]*assetpb.ResourceSearchResult, error)
	SearchAllTables() ([]*assetpb.ResourceSearchResult, error)
	SearchAllAddresses() ([]*assetpb.ResourceSearchResult, error)
	SearchAllGlobalAddresses() ([]*assetpb.Asset, error)
	SearchAllFunctions() ([]*assetpb.Asset, error)
	SearchAllSubnetworks() ([]*assetpb.ResourceSearchResult, error)
	SearchAllBigtableInstances() ([]*assetpb.Asset, error)
	SearchAllBigtableTables() ([]*assetpb.Asset, error)
	SearchAllSQLDatabaseInstances() ([]*assetpb.Asset, error)
	SearchAllHealthChecks() ([]*assetpb.ResourceSearchResult, error)
	SearchAllCloudRunServices() ([]*assetpb.ResourceSearchResult, error)
	SearchAllNodeGroups() ([]*assetpb.Asset, error)
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

func (s assetRepository) listAllResources(ty string) ([]*assetpb.Asset, error) {

	filteredResults := []*assetpb.Asset{}
	var errorString string

	for _, scope := range s.config.Scopes {
		cacheKey := fmt.Sprintf("listAllResources_%s", scope)
		cachedResults := s.cache.GetAndLock(cacheKey)
		defer s.cache.Unlock(cacheKey)

		req := &assetpb.ListAssetsRequest{
			Parent:      scope,
			ContentType: assetpb.ContentType_RESOURCE,
			AssetTypes: []string{
				cloudFunctionsFunction,
				bigtableInstanceAssetType,
				bigtableTableAssetType,
				sqlDatabaseInstanceAssetType,
				computeGlobalAddressAssetType,
				nodeGroupAssetType,
			},
		}

		var results []*assetpb.Asset

		if cachedResults != nil {
			results = cachedResults.([]*assetpb.Asset)
		}

		if results == nil {
			it := s.client.ListAssets(context.Background(), req)
			for {
				resource, err := it.Next()
				if err == iterator.Done {
					break
				}
				if err != nil && resource != nil {
					errorString = errorString + fmt.Sprintf("For scope %s on resource %s got error: %s; ", scope, resource.AssetType, err.Error())
					continue
				}
				if err != nil && resource == nil {
					return nil, err
				}
				results = append(results, resource)
			}
			s.cache.Put(cacheKey, results)
		}

		for _, result := range results {
			if result.AssetType == ty {
				filteredResults = append(filteredResults, result)
			}
		}
	}

	if len(errorString) > 0 {
		return filteredResults, errors.New(errorString)
	}

	return filteredResults, nil
}

func (s assetRepository) searchAllResources(ty string) ([]*assetpb.ResourceSearchResult, error) {

	filteredResults := []*assetpb.ResourceSearchResult{}
	var errorString string

	for _, scope := range s.config.Scopes {
		cacheKey := fmt.Sprintf("SearchAllResources_%s", scope)
		cachedResults := s.cache.GetAndLock(cacheKey)
		defer s.cache.Unlock(cacheKey)

		req := &assetpb.SearchAllResourcesRequest{
			Scope: scope,
			AssetTypes: []string{
				storageBucketAssetType,
				computeFirewallAssetType,
				computeRouterAssetType,
				computeInstanceAssetType,
				computeNetworkAssetType,
				computeSubnetworkAssetType,
				dnsManagedZoneAssetType,
				computeInstanceGroupAssetType,
				bigqueryDatasetAssetType,
				bigqueryTableAssetType,
				computeAddressAssetType,
				computeDiskAssetType,
				computeImageAssetType,
				healthCheckAssetType,
				cloudRunServiceAssetType,
			},
		}
		var results []*assetpb.ResourceSearchResult

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
				if err != nil && resource != nil {
					errorString = errorString + fmt.Sprintf("For scope %s on resource %s got error: %s; ", scope, resource.AssetType, err.Error())
					continue
				}
				if err != nil && resource == nil {
					return nil, err
				}
				results = append(results, resource)
			}
			s.cache.Put(cacheKey, results)
		}

		for _, result := range results {
			if result.AssetType == ty {
				filteredResults = append(filteredResults, result)
			}
		}
	}

	if len(errorString) > 0 {
		return filteredResults, errors.New(errorString)
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

func (s assetRepository) SearchAllGlobalAddresses() ([]*assetpb.Asset, error) {
	return s.listAllResources(computeGlobalAddressAssetType)
}

func (s assetRepository) SearchAllFunctions() ([]*assetpb.Asset, error) {
	return s.listAllResources(cloudFunctionsFunction)
}

func (s assetRepository) SearchAllSubnetworks() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(computeSubnetworkAssetType)
}

func (s assetRepository) SearchAllDisks() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(computeDiskAssetType)
}

func (s assetRepository) SearchAllImages() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(computeImageAssetType)
}

func (s assetRepository) SearchAllBigtableInstances() ([]*assetpb.Asset, error) {
	return s.listAllResources(bigtableInstanceAssetType)
}

func (s assetRepository) SearchAllBigtableTables() ([]*assetpb.Asset, error) {
	return s.listAllResources(bigtableTableAssetType)
}

func (s assetRepository) SearchAllSQLDatabaseInstances() ([]*assetpb.Asset, error) {
	return s.listAllResources(sqlDatabaseInstanceAssetType)
}

func (s assetRepository) SearchAllHealthChecks() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(healthCheckAssetType)
}

func (s assetRepository) SearchAllCloudRunServices() ([]*assetpb.ResourceSearchResult, error) {
	return s.searchAllResources(cloudRunServiceAssetType)
}

func (s assetRepository) SearchAllNodeGroups() ([]*assetpb.Asset, error) {
	return s.listAllResources(nodeGroupAssetType)
}

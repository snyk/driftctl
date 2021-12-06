package repository

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
	"github.com/snyk/driftctl/pkg/remote/azurerm/common"
	"github.com/snyk/driftctl/pkg/remote/cache"
)

type ResourcesRepository interface {
	ListAllResourceGroups() ([]*armresources.ResourceGroup, error)
}

type resourcesListPager interface {
	pager
	PageResponse() armresources.ResourceGroupsListResponse
}

type resourcesClient interface {
	List(options *armresources.ResourceGroupsListOptions) resourcesListPager
}

type resourcesClientImpl struct {
	client *armresources.ResourceGroupsClient
}

func (c resourcesClientImpl) List(options *armresources.ResourceGroupsListOptions) resourcesListPager {
	return c.client.List(options)
}

type resourcesRepository struct {
	client resourcesClient
	cache  cache.Cache
}

func NewResourcesRepository(con *arm.Connection, config common.AzureProviderConfig, cache cache.Cache) *resourcesRepository {
	return &resourcesRepository{
		&resourcesClientImpl{armresources.NewResourceGroupsClient(con, config.SubscriptionID)},
		cache,
	}
}

func (s *resourcesRepository) ListAllResourceGroups() ([]*armresources.ResourceGroup, error) {
	cacheKey := "resourcesListAllResourceGroups"
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armresources.ResourceGroup), nil
	}

	pager := s.client.List(nil)
	results := make([]*armresources.ResourceGroup, 0)
	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		if err := pager.Err(); err != nil {
			return nil, err
		}
		results = append(results, resp.ResourceGroupsListResult.Value...)
	}
	if err := pager.Err(); err != nil {
		return nil, err
	}

	s.cache.Put(cacheKey, results)

	return results, nil
}

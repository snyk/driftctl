package repository

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/containerregistry/armcontainerregistry"
	"github.com/snyk/driftctl/pkg/remote/azurerm/common"
	"github.com/snyk/driftctl/pkg/remote/cache"
)

type ContainerRegistryRepository interface {
	ListAllContainerRegistries() ([]*armcontainerregistry.Registry, error)
}

type registryClient interface {
	List(options *armcontainerregistry.RegistriesListOptions) registryListAllPager
}

type registryListAllPager interface {
	pager
	PageResponse() armcontainerregistry.RegistriesListResponse
}

type registryClientImpl struct {
	client *armcontainerregistry.RegistriesClient
}

func (c registryClientImpl) List(options *armcontainerregistry.RegistriesListOptions) registryListAllPager {
	return c.client.List(options)
}

type containerRegistryRepository struct {
	registryClient registryClient
	cache          cache.Cache
}

func NewContainerRegistryRepository(con *arm.Connection, config common.AzureProviderConfig, cache cache.Cache) *containerRegistryRepository {
	return &containerRegistryRepository{
		&registryClientImpl{client: armcontainerregistry.NewRegistriesClient(con, config.SubscriptionID)},
		cache,
	}
}

func (s *containerRegistryRepository) ListAllContainerRegistries() ([]*armcontainerregistry.Registry, error) {

	if v := s.cache.Get("ListAllContainerRegistries"); v != nil {
		return v.([]*armcontainerregistry.Registry), nil
	}

	pager := s.registryClient.List(nil)
	results := make([]*armcontainerregistry.Registry, 0)
	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		if err := pager.Err(); err != nil {
			return nil, err
		}
		results = append(results, resp.Value...)
	}

	if err := pager.Err(); err != nil {
		return nil, err
	}

	s.cache.Put("ListAllContainerRegistries", results)

	return results, nil
}

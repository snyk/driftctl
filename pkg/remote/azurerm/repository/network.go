package repository

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/network/armnetwork"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/common"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type NetworkRepository interface {
	ListAllVirtualNetworks() ([]*armnetwork.VirtualNetwork, error)
}

type virtualNetworksListAllPager interface {
	pager
	PageResponse() armnetwork.VirtualNetworksListAllResponse
}

type virtualNetworksClient interface {
	ListAll(options *armnetwork.VirtualNetworksListAllOptions) virtualNetworksListAllPager
}

type virtualNetworksClientImpl struct {
	client *armnetwork.VirtualNetworksClient
}

func (c virtualNetworksClientImpl) ListAll(options *armnetwork.VirtualNetworksListAllOptions) virtualNetworksListAllPager {
	return c.client.ListAll(options)
}

type networkRepository struct {
	virtualNetworksClient virtualNetworksClient
	cache                 cache.Cache
}

func NewNetworkRepository(con *arm.Connection, config common.AzureProviderConfig, cache cache.Cache) *networkRepository {
	return &networkRepository{
		&virtualNetworksClientImpl{client: armnetwork.NewVirtualNetworksClient(con, config.SubscriptionID)},
		cache,
	}
}

func (s *networkRepository) ListAllVirtualNetworks() ([]*armnetwork.VirtualNetwork, error) {

	if v := s.cache.Get("ListAllVirtualNetworks"); v != nil {
		return v.([]*armnetwork.VirtualNetwork), nil
	}

	pager := s.virtualNetworksClient.ListAll(nil)
	results := make([]*armnetwork.VirtualNetwork, 0)
	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		if err := pager.Err(); err != nil {
			return nil, err
		}
		results = append(results, resp.VirtualNetworksListAllResult.VirtualNetworkListResult.Value...)
	}

	if err := pager.Err(); err != nil {
		return nil, err
	}

	s.cache.Put("ListAllVirtualNetworks", results)

	return results, nil
}

package repository

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/common"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type NetworkRepository interface {
	ListAllVirtualNetworks() ([]*armnetwork.VirtualNetwork, error)
	ListAllRouteTables() ([]*armnetwork.RouteTable, error)
	ListAllSubnets(virtualNetwork *armnetwork.VirtualNetwork) ([]*armnetwork.Subnet, error)
	ListAllFirewalls() ([]*armnetwork.AzureFirewall, error)
}

type firewallsListAllPager interface {
	pager
	PageResponse() armnetwork.AzureFirewallsListAllResponse
}

type firewallsClient interface {
	ListAll(options *armnetwork.AzureFirewallsListAllOptions) firewallsListAllPager
}

type firewallsClientImpl struct {
	client *armnetwork.AzureFirewallsClient
}

func (s firewallsClientImpl) ListAll(options *armnetwork.AzureFirewallsListAllOptions) firewallsListAllPager {
	return s.client.ListAll(options)
}

type subnetsListPager interface {
	pager
	PageResponse() armnetwork.SubnetsListResponse
}

type subnetsClient interface {
	List(resourceGroupName, virtualNetworkName string, options *armnetwork.SubnetsListOptions) subnetsListPager
}

type subnetsClientImpl struct {
	client *armnetwork.SubnetsClient
}

func (s subnetsClientImpl) List(resourceGroupName, virtualNetworkName string, options *armnetwork.SubnetsListOptions) subnetsListPager {
	return s.client.List(resourceGroupName, virtualNetworkName, options)
}

type virtualNetworksClient interface {
	ListAll(options *armnetwork.VirtualNetworksListAllOptions) virtualNetworksListAllPager
}

type virtualNetworksListAllPager interface {
	pager
	PageResponse() armnetwork.VirtualNetworksListAllResponse
}

type virtualNetworksClientImpl struct {
	client *armnetwork.VirtualNetworksClient
}

func (c virtualNetworksClientImpl) ListAll(options *armnetwork.VirtualNetworksListAllOptions) virtualNetworksListAllPager {
	return c.client.ListAll(options)
}

type routeTablesClient interface {
	ListAll(options *armnetwork.RouteTablesListAllOptions) routeTablesListAllPager
}

type routeTablesListAllPager interface {
	pager
	PageResponse() armnetwork.RouteTablesListAllResponse
}

type routeTablesClientImpl struct {
	client *armnetwork.RouteTablesClient
}

func (c routeTablesClientImpl) ListAll(options *armnetwork.RouteTablesListAllOptions) routeTablesListAllPager {
	return c.client.ListAll(options)
}

type networkRepository struct {
	virtualNetworksClient virtualNetworksClient
	routeTableClient      routeTablesClient
	subnetsClient         subnetsClient
	firewallsClient       firewallsClient
	cache                 cache.Cache
}

func NewNetworkRepository(con *arm.Connection, config common.AzureProviderConfig, cache cache.Cache) *networkRepository {
	return &networkRepository{
		&virtualNetworksClientImpl{client: armnetwork.NewVirtualNetworksClient(con, config.SubscriptionID)},
		&routeTablesClientImpl{client: armnetwork.NewRouteTablesClient(con, config.SubscriptionID)},
		&subnetsClientImpl{client: armnetwork.NewSubnetsClient(con, config.SubscriptionID)},
		&firewallsClientImpl{client: armnetwork.NewAzureFirewallsClient(con, config.SubscriptionID)},
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

func (s *networkRepository) ListAllRouteTables() ([]*armnetwork.RouteTable, error) {
	if v := s.cache.Get("ListAllRouteTables"); v != nil {
		return v.([]*armnetwork.RouteTable), nil
	}

	pager := s.routeTableClient.ListAll(nil)
	results := make([]*armnetwork.RouteTable, 0)
	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		if err := pager.Err(); err != nil {
			return nil, err
		}
		results = append(results, resp.RouteTablesListAllResult.RouteTableListResult.Value...)
	}

	if err := pager.Err(); err != nil {
		return nil, err
	}

	s.cache.Put("ListAllRouteTables", results)

	return results, nil
}

func (s *networkRepository) ListAllSubnets(virtualNetwork *armnetwork.VirtualNetwork) ([]*armnetwork.Subnet, error) {

	cacheKey := fmt.Sprintf("ListAllSubnets_%s", *virtualNetwork.ID)

	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armnetwork.Subnet), nil
	}

	res, err := azure.ParseResourceID(*virtualNetwork.ID)
	if err != nil {
		return nil, err
	}

	pager := s.subnetsClient.List(res.ResourceGroup, *virtualNetwork.Name, nil)
	results := make([]*armnetwork.Subnet, 0)
	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		if err := pager.Err(); err != nil {
			return nil, err
		}
		results = append(results, resp.SubnetsListResult.SubnetListResult.Value...)
	}

	if err := pager.Err(); err != nil {
		return nil, err
	}

	s.cache.Put(cacheKey, results)

	return results, nil
}

func (s *networkRepository) ListAllFirewalls() ([]*armnetwork.AzureFirewall, error) {

	cacheKey := "ListAllFirewalls"

	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armnetwork.AzureFirewall), nil
	}

	pager := s.firewallsClient.ListAll(nil)
	results := make([]*armnetwork.AzureFirewall, 0)
	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		if err := pager.Err(); err != nil {
			return nil, err
		}
		results = append(results, resp.AzureFirewallsListAllResult.AzureFirewallListResult.Value...)
	}

	if err := pager.Err(); err != nil {
		return nil, err
	}

	s.cache.Put(cacheKey, results)

	return results, nil
}

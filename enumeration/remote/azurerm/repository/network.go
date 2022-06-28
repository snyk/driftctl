package repository

import (
	"context"
	"fmt"
	"github.com/snyk/driftctl/enumeration/remote/azurerm/common"
	"github.com/snyk/driftctl/enumeration/remote/cache"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/azure"
)

type NetworkRepository interface {
	ListAllVirtualNetworks() ([]*armnetwork.VirtualNetwork, error)
	ListAllRouteTables() ([]*armnetwork.RouteTable, error)
	ListAllSubnets(virtualNetwork *armnetwork.VirtualNetwork) ([]*armnetwork.Subnet, error)
	ListAllFirewalls() ([]*armnetwork.AzureFirewall, error)
	ListAllPublicIPAddresses() ([]*armnetwork.PublicIPAddress, error)
	ListAllSecurityGroups() ([]*armnetwork.NetworkSecurityGroup, error)
	ListAllLoadBalancers() ([]*armnetwork.LoadBalancer, error)
	ListLoadBalancerRules(*armnetwork.LoadBalancer) ([]*armnetwork.LoadBalancingRule, error)
}

type publicIPAddressesClient interface {
	ListAll(options *armnetwork.PublicIPAddressesListAllOptions) publicIPAddressesListAllPager
}

type publicIPAddressesListAllPager interface {
	pager
	PageResponse() armnetwork.PublicIPAddressesListAllResponse
}

type publicIPAddressesClientImpl struct {
	client *armnetwork.PublicIPAddressesClient
}

func (p publicIPAddressesClientImpl) ListAll(options *armnetwork.PublicIPAddressesListAllOptions) publicIPAddressesListAllPager {
	return p.client.ListAll(options)
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

type networkSecurityGroupsListAllPager interface {
	pager
	PageResponse() armnetwork.NetworkSecurityGroupsListAllResponse
}

type networkSecurityGroupsClient interface {
	ListAll(options *armnetwork.NetworkSecurityGroupsListAllOptions) networkSecurityGroupsListAllPager
}

type networkSecurityGroupsClientImpl struct {
	client *armnetwork.NetworkSecurityGroupsClient
}

func (s networkSecurityGroupsClientImpl) ListAll(options *armnetwork.NetworkSecurityGroupsListAllOptions) networkSecurityGroupsListAllPager {
	return s.client.ListAll(options)
}

type loadBalancersListAllPager interface {
	pager
	PageResponse() armnetwork.LoadBalancersListAllResponse
}

type loadBalancersClient interface {
	ListAll(options *armnetwork.LoadBalancersListAllOptions) loadBalancersListAllPager
}

type loadBalancersClientImpl struct {
	client *armnetwork.LoadBalancersClient
}

func (s loadBalancersClientImpl) ListAll(options *armnetwork.LoadBalancersListAllOptions) loadBalancersListAllPager {
	return s.client.ListAll(options)
}

type loadBalancerRulesListAllPager interface {
	pager
	PageResponse() armnetwork.LoadBalancerLoadBalancingRulesListResponse
}

type loadBalancerRulesClient interface {
	List(string, string, *armnetwork.LoadBalancerLoadBalancingRulesListOptions) loadBalancerRulesListAllPager
}

type loadBalancerRulesClientImpl struct {
	client *armnetwork.LoadBalancerLoadBalancingRulesClient
}

func (s loadBalancerRulesClientImpl) List(resourceGroupName string, loadBalancerName string, options *armnetwork.LoadBalancerLoadBalancingRulesListOptions) loadBalancerRulesListAllPager {
	return s.client.List(resourceGroupName, loadBalancerName, options)
}

type networkRepository struct {
	virtualNetworksClient       virtualNetworksClient
	routeTableClient            routeTablesClient
	subnetsClient               subnetsClient
	firewallsClient             firewallsClient
	publicIPAddressesClient     publicIPAddressesClient
	networkSecurityGroupsClient networkSecurityGroupsClient
	loadBalancersClient         loadBalancersClient
	loadBalancerRulesClient     loadBalancerRulesClient
	cache                       cache.Cache
}

func NewNetworkRepository(cred azcore.TokenCredential, options *arm.ClientOptions, config common.AzureProviderConfig, cache cache.Cache) *networkRepository {
	return &networkRepository{
		&virtualNetworksClientImpl{client: armnetwork.NewVirtualNetworksClient(config.SubscriptionID, cred, options)},
		&routeTablesClientImpl{client: armnetwork.NewRouteTablesClient(config.SubscriptionID, cred, options)},
		&subnetsClientImpl{client: armnetwork.NewSubnetsClient(config.SubscriptionID, cred, options)},
		&firewallsClientImpl{client: armnetwork.NewAzureFirewallsClient(config.SubscriptionID, cred, options)},
		&publicIPAddressesClientImpl{client: armnetwork.NewPublicIPAddressesClient(config.SubscriptionID, cred, options)},
		&networkSecurityGroupsClientImpl{client: armnetwork.NewNetworkSecurityGroupsClient(config.SubscriptionID, cred, options)},
		&loadBalancersClientImpl{client: armnetwork.NewLoadBalancersClient(config.SubscriptionID, cred, options)},
		&loadBalancerRulesClientImpl{armnetwork.NewLoadBalancerLoadBalancingRulesClient(config.SubscriptionID, cred, options)},
		cache,
	}
}

func (s *networkRepository) ListAllVirtualNetworks() ([]*armnetwork.VirtualNetwork, error) {

	cacheKey := "ListAllVirtualNetworks"
	v := s.cache.GetAndLock(cacheKey)
	defer s.cache.Unlock(cacheKey)
	if v != nil {
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

	s.cache.Put(cacheKey, results)

	return results, nil
}

func (s *networkRepository) ListAllRouteTables() ([]*armnetwork.RouteTable, error) {
	cacheKey := "ListAllRouteTables"
	v := s.cache.GetAndLock(cacheKey)
	defer s.cache.Unlock(cacheKey)
	if v != nil {
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

	s.cache.Put(cacheKey, results)

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

func (s *networkRepository) ListAllPublicIPAddresses() ([]*armnetwork.PublicIPAddress, error) {
	cacheKey := "ListAllPublicIPAddresses"

	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armnetwork.PublicIPAddress), nil
	}

	pager := s.publicIPAddressesClient.ListAll(nil)
	results := make([]*armnetwork.PublicIPAddress, 0)
	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		if err := pager.Err(); err != nil {
			return nil, err
		}
		results = append(results, resp.PublicIPAddressesListAllResult.PublicIPAddressListResult.Value...)
	}

	if err := pager.Err(); err != nil {
		return nil, err
	}

	s.cache.Put(cacheKey, results)

	return results, nil
}

func (s *networkRepository) ListAllSecurityGroups() ([]*armnetwork.NetworkSecurityGroup, error) {
	cacheKey := "networkListAllSecurityGroups"
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armnetwork.NetworkSecurityGroup), nil
	}

	pager := s.networkSecurityGroupsClient.ListAll(nil)
	results := make([]*armnetwork.NetworkSecurityGroup, 0)
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

	s.cache.Put(cacheKey, results)

	return results, nil
}

func (s *networkRepository) ListAllLoadBalancers() ([]*armnetwork.LoadBalancer, error) {
	cacheKey := "networkListAllLoadBalancers"
	defer s.cache.Unlock(cacheKey)
	if v := s.cache.GetAndLock(cacheKey); v != nil {
		return v.([]*armnetwork.LoadBalancer), nil
	}

	pager := s.loadBalancersClient.ListAll(nil)
	results := make([]*armnetwork.LoadBalancer, 0)
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

	s.cache.Put(cacheKey, results)
	return results, nil
}

func (s *networkRepository) ListLoadBalancerRules(loadBalancer *armnetwork.LoadBalancer) ([]*armnetwork.LoadBalancingRule, error) {
	cacheKey := fmt.Sprintf("networkListLoadBalancerRules_%s", *loadBalancer.ID)
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armnetwork.LoadBalancingRule), nil
	}

	loadBalancerResource, err := azure.ParseResourceID(*loadBalancer.ID)
	if err != nil {
		return nil, err
	}

	pager := s.loadBalancerRulesClient.List(loadBalancerResource.ResourceGroup, loadBalancerResource.ResourceName, &armnetwork.LoadBalancerLoadBalancingRulesListOptions{})
	results := make([]*armnetwork.LoadBalancingRule, 0)
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

	s.cache.Put(cacheKey, results)
	return results, nil
}

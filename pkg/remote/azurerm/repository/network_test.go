package repository

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/network/armnetwork"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_ListAllVirtualNetwork_MultiplesResults(t *testing.T) {

	expected := []*armnetwork.VirtualNetwork{
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("network1"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("network2"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("network3"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("network4"),
			},
		},
	}

	fakeClient := &mockVirtualNetworkClient{}

	mockPager := &mockVirtualNetworksListAllPager{}
	mockPager.On("Err").Return(nil).Times(3)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.VirtualNetworksListAllResponse{
		VirtualNetworksListAllResult: armnetwork.VirtualNetworksListAllResult{
			VirtualNetworkListResult: armnetwork.VirtualNetworkListResult{
				Value: []*armnetwork.VirtualNetwork{
					{
						Resource: armnetwork.Resource{
							ID: to.StringPtr("network1"),
						},
					},
					{
						Resource: armnetwork.Resource{
							ID: to.StringPtr("network2"),
						},
					},
				},
			},
		},
	}).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.VirtualNetworksListAllResponse{
		VirtualNetworksListAllResult: armnetwork.VirtualNetworksListAllResult{
			VirtualNetworkListResult: armnetwork.VirtualNetworkListResult{
				Value: []*armnetwork.VirtualNetwork{
					{
						Resource: armnetwork.Resource{
							ID: to.StringPtr("network3"),
						},
					},
					{
						Resource: armnetwork.Resource{
							ID: to.StringPtr("network4"),
						},
					},
				},
			},
		},
	}).Times(1)

	fakeClient.On("ListAll", mock.Anything).Return(mockPager)

	c := &cache.MockCache{}
	c.On("GetAndLock", "ListAllVirtualNetworks").Return(nil).Times(1)
	c.On("Unlock", "ListAllVirtualNetworks").Times(1)
	c.On("Put", "ListAllVirtualNetworks", expected).Return(true).Times(1)
	s := &networkRepository{
		virtualNetworksClient: fakeClient,
		cache:                 c,
	}
	got, err := s.ListAllVirtualNetworks()
	if err != nil {
		t.Errorf("ListAllVirtualNetworks() error = %v", err)
		return
	}

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllVirtualNetworks() got = %v, want %v", got, expected)
	}
}

func Test_ListAllVirtualNetwork_MultiplesResults_WithCache(t *testing.T) {

	expected := []*armnetwork.VirtualNetwork{
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("network3"),
			},
		},
	}

	fakeClient := &mockVirtualNetworkClient{}

	c := &cache.MockCache{}
	c.On("GetAndLock", "ListAllVirtualNetworks").Return(expected).Times(1)
	c.On("Unlock", "ListAllVirtualNetworks").Times(1)
	s := &networkRepository{
		virtualNetworksClient: fakeClient,
		cache:                 c,
	}
	got, err := s.ListAllVirtualNetworks()
	if err != nil {
		t.Errorf("ListAllVirtualNetworks() error = %v", err)
		return
	}

	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllVirtualNetworks() got = %v, want %v", got, expected)
	}
}

func Test_ListAllVirtualNetwork_Error_OnPageResponse(t *testing.T) {

	fakeClient := &mockVirtualNetworkClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockVirtualNetworksListAllPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.VirtualNetworksListAllResponse{}).Times(1)

	fakeClient.On("ListAll", mock.Anything).Return(mockPager)

	s := &networkRepository{
		virtualNetworksClient: fakeClient,
		cache:                 cache.New(0),
	}
	got, err := s.ListAllVirtualNetworks()

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

func Test_ListAllVirtualNetwork_Error(t *testing.T) {

	fakeClient := &mockVirtualNetworkClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockVirtualNetworksListAllPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)

	fakeClient.On("ListAll", mock.Anything).Return(mockPager)

	s := &networkRepository{
		virtualNetworksClient: fakeClient,
		cache:                 cache.New(0),
	}
	got, err := s.ListAllVirtualNetworks()

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

func Test_ListAllRouteTables_MultiplesResults(t *testing.T) {

	expected := []*armnetwork.RouteTable{
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("table1"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("table2"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("table3"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("table4"),
			},
		},
	}

	fakeClient := &mockRouteTablesClient{}

	mockPager := &mockRouteTablesListAllPager{}
	mockPager.On("Err").Return(nil).Times(3)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.RouteTablesListAllResponse{
		RouteTablesListAllResult: armnetwork.RouteTablesListAllResult{
			RouteTableListResult: armnetwork.RouteTableListResult{
				Value: expected[:2],
			},
		},
	}).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.RouteTablesListAllResponse{
		RouteTablesListAllResult: armnetwork.RouteTablesListAllResult{
			RouteTableListResult: armnetwork.RouteTableListResult{
				Value: expected[2:],
			},
		},
	}).Times(1)

	fakeClient.On("ListAll", mock.Anything).Return(mockPager)

	c := &cache.MockCache{}
	c.On("GetAndLock", "ListAllRouteTables").Return(nil).Times(1)
	c.On("Unlock", "ListAllRouteTables").Times(1)
	c.On("Put", "ListAllRouteTables", expected).Return(true).Times(1)
	s := &networkRepository{
		routeTableClient: fakeClient,
		cache:            c,
	}
	got, err := s.ListAllRouteTables()
	if err != nil {
		t.Errorf("ListAllRouteTables() error = %v", err)
		return
	}

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllRouteTables() got = %v, want %v", got, expected)
	}
}

func Test_ListAllRouteTables_MultiplesResults_WithCache(t *testing.T) {

	expected := []*armnetwork.RouteTable{
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("table1"),
			},
		},
	}

	fakeClient := &mockRouteTablesClient{}

	c := &cache.MockCache{}
	c.On("GetAndLock", "ListAllRouteTables").Return(expected).Times(1)
	c.On("Unlock", "ListAllRouteTables").Times(1)
	s := &networkRepository{
		routeTableClient: fakeClient,
		cache:            c,
	}
	got, err := s.ListAllRouteTables()
	if err != nil {
		t.Errorf("ListAllRouteTables() error = %v", err)
		return
	}

	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllRouteTables() got = %v, want %v", got, expected)
	}
}

func Test_ListAllRouteTables_Error_OnPageResponse(t *testing.T) {

	fakeClient := &mockRouteTablesClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockRouteTablesListAllPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.RouteTablesListAllResponse{}).Times(1)

	fakeClient.On("ListAll", mock.Anything).Return(mockPager)

	s := &networkRepository{
		routeTableClient: fakeClient,
		cache:            cache.New(0),
	}
	got, err := s.ListAllRouteTables()

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

func Test_ListAllRouteTables_Error(t *testing.T) {

	fakeClient := &mockRouteTablesClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockRouteTablesListAllPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)

	fakeClient.On("ListAll", mock.Anything).Return(mockPager)

	s := &networkRepository{
		routeTableClient: fakeClient,
		cache:            cache.New(0),
	}
	got, err := s.ListAllRouteTables()

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

func Test_ListAllSubnets_MultiplesResults(t *testing.T) {

	network := &armnetwork.VirtualNetwork{
		Resource: armnetwork.Resource{
			Name: to.StringPtr("network1"),
			ID:   to.StringPtr("/subscriptions/7bfb2c5c-0000-0000-0000-fffa356eb406/resourceGroups/test-dev/providers/Microsoft.Network/virtualNetworks/network1"),
		},
	}

	expected := []*armnetwork.Subnet{
		{
			SubResource: armnetwork.SubResource{
				ID: to.StringPtr("subnet1"),
			},
		},
		{
			SubResource: armnetwork.SubResource{
				ID: to.StringPtr("subnet2"),
			},
		},
		{
			SubResource: armnetwork.SubResource{
				ID: to.StringPtr("subnet3"),
			},
		},
		{
			SubResource: armnetwork.SubResource{
				ID: to.StringPtr("subnet4"),
			},
		},
	}

	fakeClient := &mockSubnetsClient{}

	mockPager := &mockSubnetsListPager{}
	mockPager.On("Err").Return(nil).Times(3)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.SubnetsListResponse{
		SubnetsListResult: armnetwork.SubnetsListResult{
			SubnetListResult: armnetwork.SubnetListResult{
				Value: []*armnetwork.Subnet{
					{
						SubResource: armnetwork.SubResource{
							ID: to.StringPtr("subnet1"),
						},
					},
					{
						SubResource: armnetwork.SubResource{
							ID: to.StringPtr("subnet2"),
						},
					},
				},
			},
		},
	}).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.SubnetsListResponse{
		SubnetsListResult: armnetwork.SubnetsListResult{
			SubnetListResult: armnetwork.SubnetListResult{
				Value: []*armnetwork.Subnet{
					{
						SubResource: armnetwork.SubResource{
							ID: to.StringPtr("subnet3"),
						},
					},
					{
						SubResource: armnetwork.SubResource{
							ID: to.StringPtr("subnet4"),
						},
					},
				},
			},
		},
	}).Times(1)

	fakeClient.On("List", "test-dev", "network1", mock.Anything).Return(mockPager)

	c := &cache.MockCache{}
	cacheKey := fmt.Sprintf("ListAllSubnets_%s", *network.ID)
	c.On("Get", cacheKey).Return(nil).Times(1)
	c.On("Put", cacheKey, expected).Return(true).Times(1)
	s := &networkRepository{
		subnetsClient: fakeClient,
		cache:         c,
	}
	got, err := s.ListAllSubnets(network)
	if err != nil {
		t.Errorf("ListAllSubnets() error = %v", err)
		return
	}

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllSubnets() got = %v, want %v", got, expected)
	}
}

func Test_ListAllSubnets_MultiplesResults_WithCache(t *testing.T) {

	network := &armnetwork.VirtualNetwork{
		Resource: armnetwork.Resource{
			ID: to.StringPtr("networkID"),
		},
	}

	expected := []*armnetwork.Subnet{
		{
			Name: to.StringPtr("network1"),
		},
	}
	fakeClient := &mockSubnetsClient{}

	c := &cache.MockCache{}
	c.On("Get", "ListAllSubnets_networkID").Return(expected).Times(1)
	s := &networkRepository{
		subnetsClient: fakeClient,
		cache:         c,
	}
	got, err := s.ListAllSubnets(network)
	if err != nil {
		t.Errorf("ListAllSubnets() error = %v", err)
		return
	}

	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllSubnets() got = %v, want %v", got, expected)
	}
}

func Test_ListAllSubnets_Error_OnPageResponse(t *testing.T) {

	network := &armnetwork.VirtualNetwork{
		Resource: armnetwork.Resource{
			Name: to.StringPtr("network1"),
			ID:   to.StringPtr("/subscriptions/7bfb2c5c-0000-0000-0000-fffa356eb406/resourceGroups/test-dev/providers/Microsoft.Network/virtualNetworks/network1"),
		},
	}

	fakeClient := &mockSubnetsClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockSubnetsListPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.SubnetsListResponse{}).Times(1)

	fakeClient.On("List", "test-dev", "network1", mock.Anything).Return(mockPager)

	s := &networkRepository{
		subnetsClient: fakeClient,
		cache:         cache.New(0),
	}
	got, err := s.ListAllSubnets(network)

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

func Test_ListAllSubnets_Error(t *testing.T) {

	network := &armnetwork.VirtualNetwork{
		Resource: armnetwork.Resource{
			Name: to.StringPtr("network1"),
			ID:   to.StringPtr("/subscriptions/7bfb2c5c-0000-0000-0000-fffa356eb406/resourceGroups/test-dev/providers/Microsoft.Network/virtualNetworks/network1"),
		},
	}

	fakeClient := &mockSubnetsClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockSubnetsListPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)

	fakeClient.On("List", "test-dev", "network1", mock.Anything).Return(mockPager)

	s := &networkRepository{
		subnetsClient: fakeClient,
		cache:         cache.New(0),
	}
	got, err := s.ListAllSubnets(network)

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

func Test_ListAllSubnets_ErrorOnInvalidNetworkID(t *testing.T) {

	network := &armnetwork.VirtualNetwork{
		Resource: armnetwork.Resource{
			Name: to.StringPtr("network1"),
			ID:   to.StringPtr("foobar"),
		},
	}

	fakeClient := &mockSubnetsClient{}

	expectedErr := errors.New("parsing failed for foobar. Invalid resource Id format")

	s := &networkRepository{
		subnetsClient: fakeClient,
		cache:         cache.New(0),
	}
	got, err := s.ListAllSubnets(network)

	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr.Error(), err.Error())
	assert.Nil(t, got)
}

func Test_ListAllFirewalls_MultiplesResults(t *testing.T) {

	expected := []*armnetwork.AzureFirewall{
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("firewall1"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("firewall2"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("firewall3"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("firewall4"),
			},
		},
	}

	fakeClient := &mockFirewallsClient{}

	mockPager := &mockFirewallsListAllPager{}
	mockPager.On("Err").Return(nil).Times(3)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.AzureFirewallsListAllResponse{
		AzureFirewallsListAllResult: armnetwork.AzureFirewallsListAllResult{
			AzureFirewallListResult: armnetwork.AzureFirewallListResult{
				Value: expected[:2],
			},
		},
	}).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.AzureFirewallsListAllResponse{
		AzureFirewallsListAllResult: armnetwork.AzureFirewallsListAllResult{
			AzureFirewallListResult: armnetwork.AzureFirewallListResult{
				Value: expected[2:],
			},
		},
	}).Times(1)

	fakeClient.On("ListAll", mock.Anything).Return(mockPager)

	c := &cache.MockCache{}
	c.On("Get", "ListAllFirewalls").Return(nil).Times(1)
	c.On("Put", "ListAllFirewalls", expected).Return(true).Times(1)
	s := &networkRepository{
		firewallsClient: fakeClient,
		cache:           c,
	}
	got, err := s.ListAllFirewalls()
	if err != nil {
		t.Errorf("ListAllFirewalls() error = %v", err)
		return
	}

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllFirewalls() got = %v, want %v", got, expected)
	}
}

func Test_ListAllFirewalls_MultiplesResults_WithCache(t *testing.T) {

	expected := []*armnetwork.AzureFirewall{
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("firewall1"),
			},
		},
	}

	fakeClient := &mockFirewallsClient{}

	c := &cache.MockCache{}
	c.On("Get", "ListAllFirewalls").Return(expected).Times(1)
	s := &networkRepository{
		firewallsClient: fakeClient,
		cache:           c,
	}
	got, err := s.ListAllFirewalls()
	if err != nil {
		t.Errorf("ListAllFirewalls() error = %v", err)
		return
	}

	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllFirewalls() got = %v, want %v", got, expected)
	}
}

func Test_ListAllFirewalls_Error_OnPageResponse(t *testing.T) {

	fakeClient := &mockFirewallsClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockFirewallsListAllPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.AzureFirewallsListAllResponse{}).Times(1)

	fakeClient.On("ListAll", mock.Anything).Return(mockPager)

	s := &networkRepository{
		firewallsClient: fakeClient,
		cache:           cache.New(0),
	}
	got, err := s.ListAllFirewalls()

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

func Test_ListAllFirewalls_Error(t *testing.T) {

	fakeClient := &mockFirewallsClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockFirewallsListAllPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)

	fakeClient.On("ListAll", mock.Anything).Return(mockPager)

	s := &networkRepository{
		firewallsClient: fakeClient,
		cache:           cache.New(0),
	}
	got, err := s.ListAllFirewalls()

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

func Test_ListAllPublicIPAddresses_MultiplesResults(t *testing.T) {

	expected := []*armnetwork.PublicIPAddress{
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("ip1"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("ip2"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("ip3"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("ip4"),
			},
		},
	}

	fakeClient := &mockPublicIPAddressesClient{}

	mockPager := &mockPublicIPAddressesListAllPager{}
	mockPager.On("Err").Return(nil).Times(3)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.PublicIPAddressesListAllResponse{
		PublicIPAddressesListAllResult: armnetwork.PublicIPAddressesListAllResult{
			PublicIPAddressListResult: armnetwork.PublicIPAddressListResult{
				Value: expected[:2],
			},
		},
	}).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.PublicIPAddressesListAllResponse{
		PublicIPAddressesListAllResult: armnetwork.PublicIPAddressesListAllResult{
			PublicIPAddressListResult: armnetwork.PublicIPAddressListResult{
				Value: expected[2:],
			},
		},
	}).Times(1)

	fakeClient.On("ListAll", mock.Anything).Return(mockPager)

	c := &cache.MockCache{}
	c.On("Get", "ListAllPublicIPAddresses").Return(nil).Times(1)
	c.On("Put", "ListAllPublicIPAddresses", expected).Return(true).Times(1)
	s := &networkRepository{
		publicIPAddressesClient: fakeClient,
		cache:                   c,
	}
	got, err := s.ListAllPublicIPAddresses()
	if err != nil {
		t.Errorf("ListAllPublicIPAddresses() error = %v", err)
		return
	}

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllPublicIPAddresses() got = %v, want %v", got, expected)
	}
}

func Test_ListAllPublicIPAddresses_MultiplesResults_WithCache(t *testing.T) {

	expected := []*armnetwork.PublicIPAddress{
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("ip1"),
			},
		},
	}

	fakeClient := &mockPublicIPAddressesClient{}

	c := &cache.MockCache{}
	c.On("Get", "ListAllPublicIPAddresses").Return(expected).Times(1)
	s := &networkRepository{
		publicIPAddressesClient: fakeClient,
		cache:                   c,
	}
	got, err := s.ListAllPublicIPAddresses()
	if err != nil {
		t.Errorf("ListAllPublicIPAddresses() error = %v", err)
		return
	}

	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllPublicIPAddresses() got = %v, want %v", got, expected)
	}
}

func Test_ListAllPublicIPAddresses_Error_OnPageResponse(t *testing.T) {

	fakeClient := &mockPublicIPAddressesClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockPublicIPAddressesListAllPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
	mockPager.On("PageResponse").Return(armnetwork.PublicIPAddressesListAllResponse{}).Times(1)

	fakeClient.On("ListAll", mock.Anything).Return(mockPager)

	s := &networkRepository{
		publicIPAddressesClient: fakeClient,
		cache:                   cache.New(0),
	}
	got, err := s.ListAllPublicIPAddresses()

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

func Test_ListAllPublicIPAddresses_Error(t *testing.T) {

	fakeClient := &mockPublicIPAddressesClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockPublicIPAddressesListAllPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)

	fakeClient.On("ListAll", mock.Anything).Return(mockPager)

	s := &networkRepository{
		publicIPAddressesClient: fakeClient,
		cache:                   cache.New(0),
	}
	got, err := s.ListAllPublicIPAddresses()

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

func Test_Network_ListAllSecurityGroups(t *testing.T) {
	expectedResults := []*armnetwork.NetworkSecurityGroup{
		{
			Resource: armnetwork.Resource{
				ID:   to.StringPtr("sgroup-1"),
				Name: to.StringPtr("sgroup-1"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID:   to.StringPtr("sgroup-2"),
				Name: to.StringPtr("sgroup-2"),
			},
		},
	}

	testcases := []struct {
		name     string
		mocks    func(*mockNetworkSecurityGroupsListAllPager, *cache.MockCache)
		expected []*armnetwork.NetworkSecurityGroup
		wantErr  string
	}{
		{
			name: "should return security groups",
			mocks: func(pager *mockNetworkSecurityGroupsListAllPager, mockCache *cache.MockCache) {
				pager.On("NextPage", context.Background()).Return(true).Times(1)
				pager.On("NextPage", context.Background()).Return(false).Times(1)
				pager.On("PageResponse").Return(armnetwork.NetworkSecurityGroupsListAllResponse{
					NetworkSecurityGroupsListAllResult: armnetwork.NetworkSecurityGroupsListAllResult{
						NetworkSecurityGroupListResult: armnetwork.NetworkSecurityGroupListResult{
							Value: expectedResults,
						},
					},
				}).Times(1)
				pager.On("Err").Return(nil).Times(2)

				mockCache.On("Get", "networkListAllSecurityGroups").Return(nil).Times(1)
				mockCache.On("Put", "networkListAllSecurityGroups", expectedResults).Return(false).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should hit cache and return security groups",
			mocks: func(pager *mockNetworkSecurityGroupsListAllPager, mockCache *cache.MockCache) {
				mockCache.On("Get", "networkListAllSecurityGroups").Return(expectedResults).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should return remote error",
			mocks: func(pager *mockNetworkSecurityGroupsListAllPager, mockCache *cache.MockCache) {
				pager.On("NextPage", context.Background()).Return(true).Times(1)
				pager.On("NextPage", context.Background()).Return(false).Times(1)
				pager.On("PageResponse").Return(armnetwork.NetworkSecurityGroupsListAllResponse{}).Times(1)
				pager.On("Err").Return(errors.New("remote error")).Times(1)

				mockCache.On("Get", "networkListAllSecurityGroups").Return(nil).Times(1)
			},
			wantErr: "remote error",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			fakePager := &mockNetworkSecurityGroupsListAllPager{}
			fakeClient := &mockNetworkSecurityGroupsClient{}
			mockCache := &cache.MockCache{}

			fakeClient.On("ListAll", (*armnetwork.NetworkSecurityGroupsListAllOptions)(nil)).Return(fakePager).Maybe()

			tt.mocks(fakePager, mockCache)

			s := &networkRepository{
				networkSecurityGroupsClient: fakeClient,
				cache:                       mockCache,
			}
			got, err := s.ListAllSecurityGroups()
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				assert.Nil(t, err)
			}

			fakeClient.AssertExpectations(t)
			mockCache.AssertExpectations(t)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ListAllSecurityGroups() got = %v, want %v", got, tt.expected)
			}
		})
	}
}

func Test_Network_ListAllLoadBalancers(t *testing.T) {
	expectedResults := []*armnetwork.LoadBalancer{
		{
			Resource: armnetwork.Resource{
				ID:   to.StringPtr("lb-1"),
				Name: to.StringPtr("lb-1"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID:   to.StringPtr("lb-2"),
				Name: to.StringPtr("lb-2"),
			},
		},
	}

	testcases := []struct {
		name     string
		mocks    func(*mockLoadBalancersListAllPager, *cache.MockCache)
		expected []*armnetwork.LoadBalancer
		wantErr  string
	}{
		{
			name: "should return load balancers",
			mocks: func(pager *mockLoadBalancersListAllPager, mockCache *cache.MockCache) {
				pager.On("NextPage", context.Background()).Return(true).Times(1)
				pager.On("NextPage", context.Background()).Return(false).Times(1)
				pager.On("PageResponse").Return(armnetwork.LoadBalancersListAllResponse{
					LoadBalancersListAllResult: armnetwork.LoadBalancersListAllResult{
						LoadBalancerListResult: armnetwork.LoadBalancerListResult{
							Value: expectedResults,
						},
					},
				}).Times(1)
				pager.On("Err").Return(nil).Times(2)

				mockCache.On("GetAndLock", "networkListAllLoadBalancers").Return(nil).Times(1)
				mockCache.On("Put", "networkListAllLoadBalancers", expectedResults).Return(false).Times(1)
				mockCache.On("Unlock", "networkListAllLoadBalancers").Return(nil).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should hit cache and return load balancers",
			mocks: func(pager *mockLoadBalancersListAllPager, mockCache *cache.MockCache) {
				mockCache.On("GetAndLock", "networkListAllLoadBalancers").Return(expectedResults).Times(1)
				mockCache.On("Unlock", "networkListAllLoadBalancers").Return(nil).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should return remote error",
			mocks: func(pager *mockLoadBalancersListAllPager, mockCache *cache.MockCache) {
				pager.On("NextPage", context.Background()).Return(true).Times(1)
				pager.On("NextPage", context.Background()).Return(false).Times(1)
				pager.On("PageResponse").Return(armnetwork.LoadBalancersListAllResponse{}).Times(1)
				pager.On("Err").Return(errors.New("remote error")).Times(1)

				mockCache.On("GetAndLock", "networkListAllLoadBalancers").Return(nil).Times(1)
				mockCache.On("Unlock", "networkListAllLoadBalancers").Return(nil).Times(1)
			},
			wantErr: "remote error",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			fakePager := &mockLoadBalancersListAllPager{}
			fakeClient := &mockLoadBalancersClient{}
			mockCache := &cache.MockCache{}

			fakeClient.On("ListAll", (*armnetwork.LoadBalancersListAllOptions)(nil)).Return(fakePager).Maybe()

			tt.mocks(fakePager, mockCache)

			s := &networkRepository{
				loadBalancersClient: fakeClient,
				cache:               mockCache,
			}
			got, err := s.ListAllLoadBalancers()
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				assert.Nil(t, err)
			}

			fakeClient.AssertExpectations(t)
			mockCache.AssertExpectations(t)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ListAllLoadBalancers() got = %v, want %v", got, tt.expected)
			}
		})
	}
}

func Test_Network_ListLoadBalancerRules(t *testing.T) {
	expectedResults := []*armnetwork.LoadBalancingRule{
		{
			SubResource: armnetwork.SubResource{
				ID: to.StringPtr("lbrule-1"),
			},
			Name: to.StringPtr("lbrule-1"),
		},
		{
			SubResource: armnetwork.SubResource{
				ID: to.StringPtr("lbrule-1"),
			},
			Name: to.StringPtr("lbrule-1"),
		},
	}

	testcases := []struct {
		name         string
		loadBalancer *armnetwork.LoadBalancer
		mocks        func(*mockLoadBalancerRulesClient, *mockLoadBalancerRulesListAllPager, *cache.MockCache)
		expected     []*armnetwork.LoadBalancingRule
		wantErr      string
	}{
		{
			name: "should return load balancer rules",
			loadBalancer: &armnetwork.LoadBalancer{
				Resource: armnetwork.Resource{ID: to.StringPtr("/subscriptions/xxx/resourceGroups/driftctl/providers/Microsoft.Network/loadBalancers/TestLoadBalancer/frontendIPConfigurations/PublicIPAddress")},
			},
			mocks: func(client *mockLoadBalancerRulesClient, pager *mockLoadBalancerRulesListAllPager, mockCache *cache.MockCache) {
				client.On("List", "driftctl", "PublicIPAddress", &armnetwork.LoadBalancerLoadBalancingRulesListOptions{}).Return(pager)

				pager.On("NextPage", context.Background()).Return(true).Times(1)
				pager.On("NextPage", context.Background()).Return(false).Times(1)
				pager.On("PageResponse").Return(armnetwork.LoadBalancerLoadBalancingRulesListResponse{
					LoadBalancerLoadBalancingRulesListResult: armnetwork.LoadBalancerLoadBalancingRulesListResult{
						LoadBalancerLoadBalancingRuleListResult: armnetwork.LoadBalancerLoadBalancingRuleListResult{
							Value: expectedResults,
						},
					},
				}).Times(1)
				pager.On("Err").Return(nil).Times(2)

				mockCache.On("Get", "networkListLoadBalancerRules_/subscriptions/xxx/resourceGroups/driftctl/providers/Microsoft.Network/loadBalancers/TestLoadBalancer/frontendIPConfigurations/PublicIPAddress").Return(nil).Times(1)
				mockCache.On("Put", "networkListLoadBalancerRules_/subscriptions/xxx/resourceGroups/driftctl/providers/Microsoft.Network/loadBalancers/TestLoadBalancer/frontendIPConfigurations/PublicIPAddress", expectedResults).Return(false).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should hit cache and return load balancers",
			loadBalancer: &armnetwork.LoadBalancer{
				Resource: armnetwork.Resource{ID: to.StringPtr("lb-1")},
			},
			mocks: func(client *mockLoadBalancerRulesClient, pager *mockLoadBalancerRulesListAllPager, mockCache *cache.MockCache) {
				mockCache.On("Get", "networkListLoadBalancerRules_lb-1").Return(expectedResults).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should return remote error",
			loadBalancer: &armnetwork.LoadBalancer{
				Resource: armnetwork.Resource{ID: to.StringPtr("/subscriptions/xxx/resourceGroups/driftctl/providers/Microsoft.Network/loadBalancers/TestLoadBalancer/frontendIPConfigurations/PublicIPAddress")},
			},
			mocks: func(client *mockLoadBalancerRulesClient, pager *mockLoadBalancerRulesListAllPager, mockCache *cache.MockCache) {
				client.On("List", "driftctl", "PublicIPAddress", &armnetwork.LoadBalancerLoadBalancingRulesListOptions{}).Return(pager)

				pager.On("NextPage", context.Background()).Return(true).Times(1)
				pager.On("NextPage", context.Background()).Return(false).Times(1)
				pager.On("PageResponse").Return(armnetwork.LoadBalancerLoadBalancingRulesListResponse{}).Times(1)
				pager.On("Err").Return(errors.New("remote error")).Times(1)

				mockCache.On("Get", "networkListLoadBalancerRules_/subscriptions/xxx/resourceGroups/driftctl/providers/Microsoft.Network/loadBalancers/TestLoadBalancer/frontendIPConfigurations/PublicIPAddress").Return(nil).Times(1)
			},
			wantErr: "remote error",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			fakePager := &mockLoadBalancerRulesListAllPager{}
			fakeClient := &mockLoadBalancerRulesClient{}
			mockCache := &cache.MockCache{}

			tt.mocks(fakeClient, fakePager, mockCache)

			s := &networkRepository{
				loadBalancerRulesClient: fakeClient,
				cache:                   mockCache,
			}
			got, err := s.ListLoadBalancerRules(tt.loadBalancer)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				assert.Nil(t, err)
			}

			fakeClient.AssertExpectations(t)
			mockCache.AssertExpectations(t)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ListAllLoadBalancers() got = %v, want %v", got, tt.expected)
			}
		})
	}
}

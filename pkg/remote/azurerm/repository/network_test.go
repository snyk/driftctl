package repository

import (
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
	c.On("Get", "ListAllVirtualNetworks").Return(nil).Times(1)
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
	c.On("Get", "ListAllVirtualNetworks").Return(expected).Times(1)
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

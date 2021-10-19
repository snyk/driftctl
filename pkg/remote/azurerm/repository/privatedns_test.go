package repository

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_ListAllPrivateZones_MultiplesResults(t *testing.T) {

	expected := []*armprivatedns.PrivateZone{
		{
			TrackedResource: armprivatedns.TrackedResource{
				Resource: armprivatedns.Resource{
					ID: func(s string) *string { return &s }("zone1"),
				},
			},
		},
		{
			TrackedResource: armprivatedns.TrackedResource{
				Resource: armprivatedns.Resource{
					ID: func(s string) *string { return &s }("zone2"),
				},
			},
		},
		{
			TrackedResource: armprivatedns.TrackedResource{
				Resource: armprivatedns.Resource{
					ID: func(s string) *string { return &s }("zone3"),
				},
			},
		},
		{
			TrackedResource: armprivatedns.TrackedResource{
				Resource: armprivatedns.Resource{
					ID: func(s string) *string { return &s }("zone4"),
				},
			},
		},
	}

	fakeClient := &mockPrivateZonesClient{}

	mockPager := &mockPrivateDNSZoneListPager{}
	mockPager.On("Err").Return(nil).Times(3)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
	mockPager.On("PageResponse").Return(armprivatedns.PrivateZonesListResponse{
		PrivateZonesListResult: armprivatedns.PrivateZonesListResult{
			PrivateZoneListResult: armprivatedns.PrivateZoneListResult{
				Value: []*armprivatedns.PrivateZone{
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID: func(s string) *string { return &s }("zone1"),
							},
						},
					},
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID: func(s string) *string { return &s }("zone2"),
							},
						},
					},
				},
			},
		},
	}).Times(1)
	mockPager.On("PageResponse").Return(armprivatedns.PrivateZonesListResponse{
		PrivateZonesListResult: armprivatedns.PrivateZonesListResult{
			PrivateZoneListResult: armprivatedns.PrivateZoneListResult{
				Value: []*armprivatedns.PrivateZone{
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID: func(s string) *string { return &s }("zone3"),
							},
						},
					},
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID: func(s string) *string { return &s }("zone4"),
							},
						},
					},
				},
			},
		},
	}).Times(1)

	fakeClient.On("List", mock.Anything).Return(mockPager)

	c := &cache.MockCache{}
	c.On("Get", "privateDNSListAllPrivateZones").Return(nil).Times(1)
	c.On("Put", "privateDNSListAllPrivateZones", expected).Return(true).Times(1)
	s := &privateDNSRepository{
		zoneClient: fakeClient,
		cache:      c,
	}
	got, err := s.ListAllPrivateZones()
	if err != nil {
		t.Errorf("ListAllPrivateZones() error = %v", err)
		return
	}

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllPrivateZones() got = %v, want %v", got, expected)
	}
}

func Test_ListAllPrivateZones_MultiplesResults_WithCache(t *testing.T) {

	expected := []*armprivatedns.PrivateZone{
		{
			TrackedResource: armprivatedns.TrackedResource{
				Resource: armprivatedns.Resource{
					ID: func(s string) *string { return &s }("zone1"),
				},
			},
		},
	}

	fakeClient := &mockPrivateZonesClient{}

	c := &cache.MockCache{}
	c.On("Get", "privateDNSListAllPrivateZones").Return(expected).Times(1)
	s := &privateDNSRepository{
		zoneClient: fakeClient,
		cache:      c,
	}
	got, err := s.ListAllPrivateZones()
	if err != nil {
		t.Errorf("ListAllPrivateZones() error = %v", err)
		return
	}

	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllPrivateZones() got = %v, want %v", got, expected)
	}
}

func Test_ListAllPrivateZones_Error(t *testing.T) {

	fakeClient := &mockPrivateZonesClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockPrivateDNSZoneListPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
	mockPager.On("PageResponse").Return(armprivatedns.PrivateZonesListResponse{}).Times(1)

	fakeClient.On("List", mock.Anything).Return(mockPager)

	s := &privateDNSRepository{
		zoneClient: fakeClient,
		cache:      cache.New(0),
	}
	got, err := s.ListAllPrivateZones()

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

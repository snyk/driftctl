package repository

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// region PrivateZone
func Test_ListAllPrivateZones_MultiplesResults(t *testing.T) {

	expected := []*armprivatedns.PrivateZone{
		{
			TrackedResource: armprivatedns.TrackedResource{
				Resource: armprivatedns.Resource{
					ID: to.StringPtr("zone1"),
				},
			},
		},
		{
			TrackedResource: armprivatedns.TrackedResource{
				Resource: armprivatedns.Resource{
					ID: to.StringPtr("zone2"),
				},
			},
		},
		{
			TrackedResource: armprivatedns.TrackedResource{
				Resource: armprivatedns.Resource{
					ID: to.StringPtr("zone3"),
				},
			},
		},
		{
			TrackedResource: armprivatedns.TrackedResource{
				Resource: armprivatedns.Resource{
					ID: to.StringPtr("zone4"),
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
								ID: to.StringPtr("zone1"),
							},
						},
					},
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID: to.StringPtr("zone2"),
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
								ID: to.StringPtr("zone3"),
							},
						},
					},
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID: to.StringPtr("zone4"),
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
					ID: to.StringPtr("zone1"),
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

// endregion

// region ARecord
func Test_ListAllARecords_MultiplesResults(t *testing.T) {

	expected := []*armprivatedns.RecordSet{
		{
			ProxyResource: armprivatedns.ProxyResource{
				Resource: armprivatedns.Resource{
					ID: to.StringPtr("record1"),
				},
			},
			Properties: &armprivatedns.RecordSetProperties{
				ARecords: []*armprivatedns.ARecord{
					{IPv4Address: to.StringPtr("ip")},
				},
			},
		},
		{
			ProxyResource: armprivatedns.ProxyResource{
				Resource: armprivatedns.Resource{
					ID: to.StringPtr("record3"),
				},
			},
			Properties: &armprivatedns.RecordSetProperties{
				ARecords: []*armprivatedns.ARecord{
					{IPv4Address: to.StringPtr("ip")},
				},
			},
		},
	}

	fakeRecordSetClient := &mockPrivateRecordSetClient{}

	mockPager := &mockPrivateDNSRecordSetListPager{}
	mockPager.On("Err").Return(nil).Times(3)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
	mockPager.On("PageResponse").Return(armprivatedns.RecordSetsListResponse{
		RecordSetsListResult: armprivatedns.RecordSetsListResult{
			RecordSetListResult: armprivatedns.RecordSetListResult{
				Value: []*armprivatedns.RecordSet{
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID: to.StringPtr("record1"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							ARecords: []*armprivatedns.ARecord{
								{IPv4Address: to.StringPtr("ip")},
							},
						},
					},
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID: to.StringPtr("record2"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{},
					},
				},
			},
		},
	}).Times(1)
	mockPager.On("PageResponse").Return(armprivatedns.RecordSetsListResponse{
		RecordSetsListResult: armprivatedns.RecordSetsListResult{
			RecordSetListResult: armprivatedns.RecordSetListResult{
				Value: []*armprivatedns.RecordSet{
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID: to.StringPtr("record3"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							ARecords: []*armprivatedns.ARecord{
								{IPv4Address: to.StringPtr("ip")},
							},
						},
					},
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID: to.StringPtr("record4"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{},
					},
				},
			},
		},
	}).Times(1)

	fakeRecordSetClient.On("List", "rgid", "zone", (*armprivatedns.RecordSetsListOptions)(nil)).Return(mockPager)

	c := &cache.MockCache{}
	c.On("Get", "privateDNSListAllARecords-/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com").Return(nil).Times(1)
	c.On("Put", "privateDNSListAllARecords-/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com", expected).Return(true).Times(1)
	c.On("GetAndLock", "privateDNSlistAllRecords-/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com").Return(nil).Times(1)
	c.On("Unlock", "privateDNSlistAllRecords-/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com").Return().Times(1)
	c.On("Put", "privateDNSlistAllRecords-/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com", mock.Anything).Return(true).Times(1)
	s := &privateDNSRepository{
		recordClient: fakeRecordSetClient,
		cache:        c,
	}
	got, err := s.ListAllARecords(&armprivatedns.PrivateZone{
		TrackedResource: armprivatedns.TrackedResource{
			Resource: armprivatedns.Resource{
				ID:   to.StringPtr("/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com"),
				Name: to.StringPtr("zone"),
			},
		},
	})
	if err != nil {
		t.Errorf("ListAllARecords() error = %v", err)
		return
	}

	mockPager.AssertExpectations(t)
	fakeRecordSetClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllARecords() got = %v, want %v", got, expected)
	}
}

func Test_ListAllARecords_MultiplesResults_WithCache(t *testing.T) {

	expected := []*armprivatedns.RecordSet{
		{
			ProxyResource: armprivatedns.ProxyResource{
				Resource: armprivatedns.Resource{
					ID: to.StringPtr("record1"),
				},
			},
		},
	}

	fakeRecordSetClient := &mockPrivateRecordSetClient{}

	c := &cache.MockCache{}
	c.On("Get", "privateDNSListAllARecords-/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com").Return(expected).Times(1)
	s := &privateDNSRepository{
		recordClient: fakeRecordSetClient,
		cache:        c,
	}
	got, err := s.ListAllARecords(&armprivatedns.PrivateZone{
		TrackedResource: armprivatedns.TrackedResource{
			Resource: armprivatedns.Resource{
				ID:   to.StringPtr("/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com"),
				Name: to.StringPtr("zone"),
			},
		},
	})
	if err != nil {
		t.Errorf("ListAllARecords() error = %v", err)
		return
	}

	fakeRecordSetClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllARecords() got = %v, want %v", got, expected)
	}
}

func Test_ListAllARecords_Error(t *testing.T) {

	fakeClient := &mockPrivateRecordSetClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockPrivateDNSRecordSetListPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
	mockPager.On("PageResponse").Return(armprivatedns.RecordSetsListResponse{}).Times(1)

	fakeClient.On("List", "rgid", "zone", (*armprivatedns.RecordSetsListOptions)(nil)).Return(mockPager)

	s := &privateDNSRepository{
		recordClient: fakeClient,
		cache:        cache.New(0),
	}
	got, err := s.ListAllARecords(&armprivatedns.PrivateZone{
		TrackedResource: armprivatedns.TrackedResource{
			Resource: armprivatedns.Resource{
				ID:   to.StringPtr("/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com"),
				Name: to.StringPtr("zone"),
			},
		},
	})

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

// endregion

// region AAAAAAARecord
func Test_ListAllAAAARecords_MultiplesResults(t *testing.T) {

	expected := []*armprivatedns.RecordSet{
		{
			ProxyResource: armprivatedns.ProxyResource{
				Resource: armprivatedns.Resource{
					ID: to.StringPtr("record1"),
				},
			},
			Properties: &armprivatedns.RecordSetProperties{
				AaaaRecords: []*armprivatedns.AaaaRecord{
					{IPv6Address: to.StringPtr("ip")},
				},
			},
		},
		{
			ProxyResource: armprivatedns.ProxyResource{
				Resource: armprivatedns.Resource{
					ID: to.StringPtr("record3"),
				},
			},
			Properties: &armprivatedns.RecordSetProperties{
				AaaaRecords: []*armprivatedns.AaaaRecord{
					{IPv6Address: to.StringPtr("ip")},
				},
			},
		},
	}

	fakeRecordSetClient := &mockPrivateRecordSetClient{}

	mockPager := &mockPrivateDNSRecordSetListPager{}
	mockPager.On("Err").Return(nil).Times(3)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
	mockPager.On("PageResponse").Return(armprivatedns.RecordSetsListResponse{
		RecordSetsListResult: armprivatedns.RecordSetsListResult{
			RecordSetListResult: armprivatedns.RecordSetListResult{
				Value: []*armprivatedns.RecordSet{
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID: to.StringPtr("record1"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							AaaaRecords: []*armprivatedns.AaaaRecord{
								{IPv6Address: to.StringPtr("ip")},
							},
						},
					},
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID: to.StringPtr("record2"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{},
					},
				},
			},
		},
	}).Times(1)
	mockPager.On("PageResponse").Return(armprivatedns.RecordSetsListResponse{
		RecordSetsListResult: armprivatedns.RecordSetsListResult{
			RecordSetListResult: armprivatedns.RecordSetListResult{
				Value: []*armprivatedns.RecordSet{
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID: to.StringPtr("record3"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							AaaaRecords: []*armprivatedns.AaaaRecord{
								{IPv6Address: to.StringPtr("ip")},
							},
						},
					},
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID: to.StringPtr("record4"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{},
					},
				},
			},
		},
	}).Times(1)

	fakeRecordSetClient.On("List", "rgid", "zone", (*armprivatedns.RecordSetsListOptions)(nil)).Return(mockPager)

	c := &cache.MockCache{}
	c.On("Get", "privateDNSListAllAAAARecords-/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com").Return(nil).Times(1)
	c.On("Put", "privateDNSListAllAAAARecords-/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com", expected).Return(true).Times(1)
	c.On("GetAndLock", "privateDNSlistAllRecords-/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com").Return(nil).Times(1)
	c.On("Unlock", "privateDNSlistAllRecords-/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com").Return().Times(1)
	c.On("Put", "privateDNSlistAllRecords-/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com", mock.Anything).Return(true).Times(1)
	s := &privateDNSRepository{
		recordClient: fakeRecordSetClient,
		cache:        c,
	}
	got, err := s.ListAllAAAARecords(&armprivatedns.PrivateZone{
		TrackedResource: armprivatedns.TrackedResource{
			Resource: armprivatedns.Resource{
				ID:   to.StringPtr("/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com"),
				Name: to.StringPtr("zone"),
			},
		},
	})
	if err != nil {
		t.Errorf("ListAllAAAARecords() error = %v", err)
		return
	}

	mockPager.AssertExpectations(t)
	fakeRecordSetClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllAAAARecords() got = %v, want %v", got, expected)
	}
}

func Test_ListAllAAAARecords_MultiplesResults_WithCache(t *testing.T) {

	expected := []*armprivatedns.RecordSet{
		{
			ProxyResource: armprivatedns.ProxyResource{
				Resource: armprivatedns.Resource{
					ID: to.StringPtr("record1"),
				},
			},
		},
	}

	fakeRecordSetClient := &mockPrivateRecordSetClient{}

	c := &cache.MockCache{}
	c.On("Get", "privateDNSListAllAAAARecords-/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com").Return(expected).Times(1)
	s := &privateDNSRepository{
		recordClient: fakeRecordSetClient,
		cache:        c,
	}
	got, err := s.ListAllAAAARecords(&armprivatedns.PrivateZone{
		TrackedResource: armprivatedns.TrackedResource{
			Resource: armprivatedns.Resource{
				ID:   to.StringPtr("/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com"),
				Name: to.StringPtr("zone"),
			},
		},
	})
	if err != nil {
		t.Errorf("ListAllAAAARecords() error = %v", err)
		return
	}

	fakeRecordSetClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllAAAARecords() got = %v, want %v", got, expected)
	}
}

func Test_ListAllAAAARecords_Error(t *testing.T) {

	fakeClient := &mockPrivateRecordSetClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockPrivateDNSRecordSetListPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
	mockPager.On("PageResponse").Return(armprivatedns.RecordSetsListResponse{}).Times(1)

	fakeClient.On("List", "rgid", "zone", (*armprivatedns.RecordSetsListOptions)(nil)).Return(mockPager)

	s := &privateDNSRepository{
		recordClient: fakeClient,
		cache:        cache.New(0),
	}
	got, err := s.ListAllAAAARecords(&armprivatedns.PrivateZone{
		TrackedResource: armprivatedns.TrackedResource{
			Resource: armprivatedns.Resource{
				ID:   to.StringPtr("/subscriptions/subid/resourceGroups/rgid/providers/Microsoft.Network/privateDnsZones/zone.com"),
				Name: to.StringPtr("zone"),
			},
		},
	})

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

// endregion

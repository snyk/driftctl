// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package repository

import (
	armprivatedns "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	mock "github.com/stretchr/testify/mock"
)

// MockPrivateDNSRepository is an autogenerated mock type for the PrivateDNSRepository type
type MockPrivateDNSRepository struct {
	mock.Mock
}

// ListAllAAAARecords provides a mock function with given fields: zone
func (_m *MockPrivateDNSRepository) ListAllAAAARecords(zone *armprivatedns.PrivateZone) ([]*armprivatedns.RecordSet, error) {
	ret := _m.Called(zone)

	var r0 []*armprivatedns.RecordSet
	if rf, ok := ret.Get(0).(func(*armprivatedns.PrivateZone) []*armprivatedns.RecordSet); ok {
		r0 = rf(zone)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*armprivatedns.RecordSet)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*armprivatedns.PrivateZone) error); ok {
		r1 = rf(zone)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllARecords provides a mock function with given fields: zone
func (_m *MockPrivateDNSRepository) ListAllARecords(zone *armprivatedns.PrivateZone) ([]*armprivatedns.RecordSet, error) {
	ret := _m.Called(zone)

	var r0 []*armprivatedns.RecordSet
	if rf, ok := ret.Get(0).(func(*armprivatedns.PrivateZone) []*armprivatedns.RecordSet); ok {
		r0 = rf(zone)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*armprivatedns.RecordSet)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*armprivatedns.PrivateZone) error); ok {
		r1 = rf(zone)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllCNAMERecords provides a mock function with given fields: zone
func (_m *MockPrivateDNSRepository) ListAllCNAMERecords(zone *armprivatedns.PrivateZone) ([]*armprivatedns.RecordSet, error) {
	ret := _m.Called(zone)

	var r0 []*armprivatedns.RecordSet
	if rf, ok := ret.Get(0).(func(*armprivatedns.PrivateZone) []*armprivatedns.RecordSet); ok {
		r0 = rf(zone)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*armprivatedns.RecordSet)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*armprivatedns.PrivateZone) error); ok {
		r1 = rf(zone)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllMXRecords provides a mock function with given fields: zone
func (_m *MockPrivateDNSRepository) ListAllMXRecords(zone *armprivatedns.PrivateZone) ([]*armprivatedns.RecordSet, error) {
	ret := _m.Called(zone)

	var r0 []*armprivatedns.RecordSet
	if rf, ok := ret.Get(0).(func(*armprivatedns.PrivateZone) []*armprivatedns.RecordSet); ok {
		r0 = rf(zone)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*armprivatedns.RecordSet)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*armprivatedns.PrivateZone) error); ok {
		r1 = rf(zone)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllPTRRecords provides a mock function with given fields: zone
func (_m *MockPrivateDNSRepository) ListAllPTRRecords(zone *armprivatedns.PrivateZone) ([]*armprivatedns.RecordSet, error) {
	ret := _m.Called(zone)

	var r0 []*armprivatedns.RecordSet
	if rf, ok := ret.Get(0).(func(*armprivatedns.PrivateZone) []*armprivatedns.RecordSet); ok {
		r0 = rf(zone)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*armprivatedns.RecordSet)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*armprivatedns.PrivateZone) error); ok {
		r1 = rf(zone)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllPrivateZones provides a mock function with given fields:
func (_m *MockPrivateDNSRepository) ListAllPrivateZones() ([]*armprivatedns.PrivateZone, error) {
	ret := _m.Called()

	var r0 []*armprivatedns.PrivateZone
	if rf, ok := ret.Get(0).(func() []*armprivatedns.PrivateZone); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*armprivatedns.PrivateZone)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllSRVRecords provides a mock function with given fields: zone
func (_m *MockPrivateDNSRepository) ListAllSRVRecords(zone *armprivatedns.PrivateZone) ([]*armprivatedns.RecordSet, error) {
	ret := _m.Called(zone)

	var r0 []*armprivatedns.RecordSet
	if rf, ok := ret.Get(0).(func(*armprivatedns.PrivateZone) []*armprivatedns.RecordSet); ok {
		r0 = rf(zone)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*armprivatedns.RecordSet)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*armprivatedns.PrivateZone) error); ok {
		r1 = rf(zone)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllTXTRecords provides a mock function with given fields: zone
func (_m *MockPrivateDNSRepository) ListAllTXTRecords(zone *armprivatedns.PrivateZone) ([]*armprivatedns.RecordSet, error) {
	ret := _m.Called(zone)

	var r0 []*armprivatedns.RecordSet
	if rf, ok := ret.Get(0).(func(*armprivatedns.PrivateZone) []*armprivatedns.RecordSet); ok {
		r0 = rf(zone)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*armprivatedns.RecordSet)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*armprivatedns.PrivateZone) error); ok {
		r1 = rf(zone)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package repository

import mock "github.com/stretchr/testify/mock"

// MockCloudResourceManagerRepository is an autogenerated mock type for the CloudResourceManagerRepository type
type MockCloudResourceManagerRepository struct {
	mock.Mock
}

// ListProjectsBindings provides a mock function with given fields:
func (_m *MockCloudResourceManagerRepository) ListProjectsBindings() (map[string]map[string][]string, error) {
	ret := _m.Called()

	var r0 map[string]map[string][]string
	if rf, ok := ret.Get(0).(func() map[string]map[string][]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]map[string][]string)
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
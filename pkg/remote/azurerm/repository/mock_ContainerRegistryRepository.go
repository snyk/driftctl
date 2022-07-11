// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package repository

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	mock "github.com/stretchr/testify/mock"
)

// MockContainerRegistryRepository is an autogenerated mock type for the ContainerRegistryRepository type
type MockContainerRegistryRepository struct {
	mock.Mock
}

// ListAllContainerRegistries provides a mock function with given fields:
func (_m *MockContainerRegistryRepository) ListAllContainerRegistries() ([]*armcontainerregistry.Registry, error) {
	ret := _m.Called()

	var r0 []*armcontainerregistry.Registry
	if rf, ok := ret.Get(0).(func() []*armcontainerregistry.Registry); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*armcontainerregistry.Registry)
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
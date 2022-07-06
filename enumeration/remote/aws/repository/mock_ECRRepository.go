// Code generated by mockery v2.10.0. DO NOT EDIT.

package repository

import (
	ecr "github.com/aws/aws-sdk-go/service/ecr"
	mock "github.com/stretchr/testify/mock"
)

// MockECRRepository is an autogenerated mock type for the ECRRepository type
type MockECRRepository struct {
	mock.Mock
}

// GetRepositoryPolicy provides a mock function with given fields: _a0
func (_m *MockECRRepository) GetRepositoryPolicy(_a0 *ecr.Repository) (*ecr.GetRepositoryPolicyOutput, error) {
	ret := _m.Called(_a0)

	var r0 *ecr.GetRepositoryPolicyOutput
	if rf, ok := ret.Get(0).(func(*ecr.Repository) *ecr.GetRepositoryPolicyOutput); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ecr.GetRepositoryPolicyOutput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*ecr.Repository) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllRepositories provides a mock function with given fields:
func (_m *MockECRRepository) ListAllRepositories() ([]*ecr.Repository, error) {
	ret := _m.Called()

	var r0 []*ecr.Repository
	if rf, ok := ret.Get(0).(func() []*ecr.Repository); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*ecr.Repository)
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
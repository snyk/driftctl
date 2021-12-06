package repository

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/containerregistry/armcontainerregistry"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Resources_ListAllContainerRegistries(t *testing.T) {
	expectedResults := []*armcontainerregistry.Registry{
		{
			Resource: armcontainerregistry.Resource{
				ID:   to.StringPtr("/subscriptions/2c361f34-30fb-47ae-a227-83a5d3a26c66/resourceGroups/my-group/providers/Microsoft.ContainerRegistry/registries/containerRegistry1"),
				Name: to.StringPtr("containerRegistry1"),
			},
		},
		{
			Resource: armcontainerregistry.Resource{
				ID:   to.StringPtr("/subscriptions/2c361f34-30fb-47ae-a227-83a5d3a26c66/resourceGroups/my-group/providers/Microsoft.ContainerRegistry/registries/containerRegistry1"),
				Name: to.StringPtr("containerRegistry2"),
			},
		},
		{
			Resource: armcontainerregistry.Resource{
				ID:   to.StringPtr("/subscriptions/008b5f48-1b66-4d92-a6b6-d215b4c9b473/-/resource-3"),
				Name: to.StringPtr("resource-3"),
			},
		},
	}

	testcases := []struct {
		name     string
		mocks    func(*mockRegistryListAllPager, *cache.MockCache)
		expected []*armcontainerregistry.Registry
		wantErr  string
	}{
		{
			name: "should return container registries",
			mocks: func(mockPager *mockRegistryListAllPager, mockCache *cache.MockCache) {
				mockPager.On("Err").Return(nil).Times(3)
				mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
				mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
				mockPager.On("PageResponse").Return(armcontainerregistry.RegistriesListResponse{
					RegistriesListResult: armcontainerregistry.RegistriesListResult{
						RegistryListResult: armcontainerregistry.RegistryListResult{
							Value: expectedResults[:2],
						},
					},
				}).Times(1)
				mockPager.On("PageResponse").Return(armcontainerregistry.RegistriesListResponse{
					RegistriesListResult: armcontainerregistry.RegistriesListResult{
						RegistryListResult: armcontainerregistry.RegistryListResult{
							Value: expectedResults[2:],
						},
					},
				}).Times(1)

				mockCache.On("Get", "ListAllContainerRegistries").Return(nil).Times(1)
				mockCache.On("Put", "ListAllContainerRegistries", expectedResults).Return(false).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should hit cache and return container registries",
			mocks: func(mockPager *mockRegistryListAllPager, mockCache *cache.MockCache) {
				mockCache.On("Get", "ListAllContainerRegistries").Return(expectedResults).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should return remote error",
			mocks: func(mockPager *mockRegistryListAllPager, mockCache *cache.MockCache) {
				mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
				mockPager.On("PageResponse").Return(armcontainerregistry.RegistriesListResponse{
					RegistriesListResult: armcontainerregistry.RegistriesListResult{
						RegistryListResult: armcontainerregistry.RegistryListResult{
							Value: []*armcontainerregistry.Registry{},
						},
					},
				}).Times(1)
				mockPager.On("Err").Return(errors.New("remote error")).Times(1)

				mockCache.On("Get", "ListAllContainerRegistries").Return(nil).Times(1)
			},
			wantErr: "remote error",
		},
		{
			name: "should return remote error after fetching all pages",
			mocks: func(mockPager *mockRegistryListAllPager, mockCache *cache.MockCache) {
				mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
				mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
				mockPager.On("PageResponse").Return(armcontainerregistry.RegistriesListResponse{
					RegistriesListResult: armcontainerregistry.RegistriesListResult{
						RegistryListResult: armcontainerregistry.RegistryListResult{
							Value: []*armcontainerregistry.Registry{},
						},
					},
				}).Times(1)
				mockPager.On("Err").Return(nil).Times(1)
				mockPager.On("Err").Return(errors.New("remote error")).Times(1)

				mockCache.On("Get", "ListAllContainerRegistries").Return(nil).Times(1)
			},
			wantErr: "remote error",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := &mockRegistryClient{}
			mockPager := &mockRegistryListAllPager{}
			mockCache := &cache.MockCache{}

			fakeClient.On("List", mock.Anything).Maybe().Return(mockPager)

			tt.mocks(mockPager, mockCache)

			s := &containerRegistryRepository{
				registryClient: fakeClient,
				cache:          mockCache,
			}
			got, err := s.ListAllContainerRegistries()
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
			} else {
				assert.Nil(t, err)
			}

			fakeClient.AssertExpectations(t)
			mockPager.AssertExpectations(t)
			mockCache.AssertExpectations(t)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ListAllResourceGroups() got = %v, want %v", got, tt.expected)
			}
		})
	}
}

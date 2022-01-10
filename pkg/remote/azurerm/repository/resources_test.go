package repository

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Resources_ListAllResourceGroups(t *testing.T) {
	expectedResults := []*armresources.ResourceGroup{
		{
			ID:   to.StringPtr("/subscriptions/008b5f48-1b66-4d92-a6b6-d215b4c9b473/resourceGroups/elie-dev"),
			Name: to.StringPtr("elie-dev"),
		},
		{
			ID:   to.StringPtr("/subscriptions/008b5f48-1b66-4d92-a6b6-d215b4c9b473/resourceGroups/william-dev"),
			Name: to.StringPtr("william-dev"),
		},
		{
			ID:   to.StringPtr("/subscriptions/008b5f48-1b66-4d92-a6b6-d215b4c9b473/resourceGroups/driftctl-sj-tests"),
			Name: to.StringPtr("driftctl-sj-tests"),
		},
	}

	testcases := []struct {
		name     string
		mocks    func(*mockResourcesListPager, *cache.MockCache)
		expected []*armresources.ResourceGroup
		wantErr  string
	}{
		{
			name: "should return resource groups",
			mocks: func(mockPager *mockResourcesListPager, mockCache *cache.MockCache) {
				mockPager.On("Err").Return(nil).Times(3)
				mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
				mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
				mockPager.On("PageResponse").Return(armresources.ResourceGroupsListResponse{
					ResourceGroupsListResult: armresources.ResourceGroupsListResult{
						ResourceGroupListResult: armresources.ResourceGroupListResult{
							Value: []*armresources.ResourceGroup{
								{
									ID:   to.StringPtr("/subscriptions/008b5f48-1b66-4d92-a6b6-d215b4c9b473/resourceGroups/elie-dev"),
									Name: to.StringPtr("elie-dev"),
								},
								{
									ID:   to.StringPtr("/subscriptions/008b5f48-1b66-4d92-a6b6-d215b4c9b473/resourceGroups/william-dev"),
									Name: to.StringPtr("william-dev"),
								},
							},
						},
					},
				}).Times(1)
				mockPager.On("PageResponse").Return(armresources.ResourceGroupsListResponse{
					ResourceGroupsListResult: armresources.ResourceGroupsListResult{
						ResourceGroupListResult: armresources.ResourceGroupListResult{
							Value: []*armresources.ResourceGroup{
								{
									ID:   to.StringPtr("/subscriptions/008b5f48-1b66-4d92-a6b6-d215b4c9b473/resourceGroups/driftctl-sj-tests"),
									Name: to.StringPtr("driftctl-sj-tests"),
								},
							},
						},
					},
				}).Times(1)

				mockCache.On("Get", "resourcesListAllResourceGroups").Return(nil).Times(1)
				mockCache.On("Put", "resourcesListAllResourceGroups", expectedResults).Return(true).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should hit cache and return resource groups",
			mocks: func(mockPager *mockResourcesListPager, mockCache *cache.MockCache) {
				mockCache.On("Get", "resourcesListAllResourceGroups").Return(expectedResults).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should return remote error",
			mocks: func(mockPager *mockResourcesListPager, mockCache *cache.MockCache) {
				mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
				mockPager.On("PageResponse").Return(armresources.ResourceGroupsListResponse{
					ResourceGroupsListResult: armresources.ResourceGroupsListResult{
						ResourceGroupListResult: armresources.ResourceGroupListResult{
							Value: []*armresources.ResourceGroup{},
						},
					},
				}).Times(1)
				mockPager.On("Err").Return(errors.New("remote error")).Times(1)

				mockCache.On("Get", "resourcesListAllResourceGroups").Return(nil).Times(1)
			},
			wantErr: "remote error",
		},
		{
			name: "should return remote error after fetching all pages",
			mocks: func(mockPager *mockResourcesListPager, mockCache *cache.MockCache) {
				mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
				mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
				mockPager.On("PageResponse").Return(armresources.ResourceGroupsListResponse{
					ResourceGroupsListResult: armresources.ResourceGroupsListResult{
						ResourceGroupListResult: armresources.ResourceGroupListResult{
							Value: []*armresources.ResourceGroup{},
						},
					},
				}).Times(1)
				mockPager.On("Err").Return(nil).Times(1)
				mockPager.On("Err").Return(errors.New("remote error")).Times(1)

				mockCache.On("Get", "resourcesListAllResourceGroups").Return(nil).Times(1)
			},
			wantErr: "remote error",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := &mockResourcesClient{}
			mockPager := &mockResourcesListPager{}
			mockCache := &cache.MockCache{}

			fakeClient.On("List", mock.Anything).Maybe().Return(mockPager)

			tt.mocks(mockPager, mockCache)

			s := &resourcesRepository{
				client: fakeClient,
				cache:  mockCache,
			}
			got, err := s.ListAllResourceGroups()
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

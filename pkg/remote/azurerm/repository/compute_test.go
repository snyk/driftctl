package repository

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Compute_ListAllImages(t *testing.T) {
	expectedResults := []*armcompute.Image{
		{
			Resource: armcompute.Resource{
				ID:   to.StringPtr("/subscriptions/2c361f34-30fb-47ae-a227-83a5d3a26c66/resourceGroups/tfvmex-resources/providers/Microsoft.Compute/images/image1"),
				Name: to.StringPtr("image1"),
			},
		},
		{
			Resource: armcompute.Resource{
				ID:   to.StringPtr("/subscriptions/2c361f34-30fb-47ae-a227-83a5d3a26c66/resourceGroups/tfvmex-resources/providers/Microsoft.Compute/images/image2"),
				Name: to.StringPtr("image2"),
			},
		},
		{
			Resource: armcompute.Resource{
				ID:   to.StringPtr("/subscriptions/2c361f34-30fb-47ae-a227-83a5d3a26c66/resourceGroups/tfvmex-resources/providers/Microsoft.Compute/images/image3"),
				Name: to.StringPtr("image3"),
			},
		},
	}

	testcases := []struct {
		name     string
		mocks    func(*mockImagesListPager, *cache.MockCache)
		expected []*armcompute.Image
		wantErr  string
	}{
		{
			name: "should return images",
			mocks: func(mockPager *mockImagesListPager, mockCache *cache.MockCache) {
				mockPager.On("Err").Return(nil).Times(3)
				mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
				mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
				mockPager.On("PageResponse").Return(armcompute.ImagesListResponse{
					ImagesListResult: armcompute.ImagesListResult{
						ImageListResult: armcompute.ImageListResult{
							Value: expectedResults[:2],
						},
					},
				}).Times(1)
				mockPager.On("PageResponse").Return(armcompute.ImagesListResponse{
					ImagesListResult: armcompute.ImagesListResult{
						ImageListResult: armcompute.ImageListResult{
							Value: expectedResults[2:],
						},
					},
				}).Times(1)

				mockCache.On("Get", "computeListAllImages").Return(nil).Times(1)
				mockCache.On("Put", "computeListAllImages", expectedResults).Return(false).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should hit cache and return images",
			mocks: func(mockPager *mockImagesListPager, mockCache *cache.MockCache) {
				mockCache.On("Get", "computeListAllImages").Return(expectedResults).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should return remote error",
			mocks: func(mockPager *mockImagesListPager, mockCache *cache.MockCache) {
				mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
				mockPager.On("PageResponse").Return(armcompute.ImagesListResponse{
					ImagesListResult: armcompute.ImagesListResult{
						ImageListResult: armcompute.ImageListResult{
							Value: []*armcompute.Image{},
						},
					},
				}).Times(1)
				mockPager.On("Err").Return(errors.New("remote error")).Times(1)

				mockCache.On("Get", "computeListAllImages").Return(nil).Times(1)
			},
			wantErr: "remote error",
		},
		{
			name: "should return remote error after fetching all pages",
			mocks: func(mockPager *mockImagesListPager, mockCache *cache.MockCache) {
				mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
				mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
				mockPager.On("PageResponse").Return(armcompute.ImagesListResponse{
					ImagesListResult: armcompute.ImagesListResult{
						ImageListResult: armcompute.ImageListResult{
							Value: []*armcompute.Image{},
						},
					},
				}).Times(1)
				mockPager.On("Err").Return(nil).Times(1)
				mockPager.On("Err").Return(errors.New("remote error")).Times(1)

				mockCache.On("Get", "computeListAllImages").Return(nil).Times(1)
			},
			wantErr: "remote error",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := &mockImagesClient{}
			mockPager := &mockImagesListPager{}
			mockCache := &cache.MockCache{}

			fakeClient.On("List", mock.Anything).Maybe().Return(mockPager)

			tt.mocks(mockPager, mockCache)

			s := &computeRepository{
				imagesClient: fakeClient,
				cache:        mockCache,
			}
			got, err := s.ListAllImages()
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

func Test_Compute_ListAllSSHPublicKeys(t *testing.T) {
	expectedResults := []*armcompute.SSHPublicKeyResource{
		{
			Resource: armcompute.Resource{
				ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/TESTRESGROUP/providers/Microsoft.Compute/sshPublicKeys/key1"),
				Name: to.StringPtr("key1"),
			},
		},
		{
			Resource: armcompute.Resource{
				ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/TESTRESGROUP/providers/Microsoft.Compute/sshPublicKeys/key2"),
				Name: to.StringPtr("key2"),
			},
		},
		{
			Resource: armcompute.Resource{
				ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/TESTRESGROUP/providers/Microsoft.Compute/sshPublicKeys/key3"),
				Name: to.StringPtr("key3"),
			},
		},
	}

	testcases := []struct {
		name     string
		mocks    func(*mockSshPublicKeyListPager, *cache.MockCache)
		expected []*armcompute.SSHPublicKeyResource
		wantErr  string
	}{
		{
			name: "should return SSH public keys",
			mocks: func(mockPager *mockSshPublicKeyListPager, mockCache *cache.MockCache) {
				mockPager.On("Err").Return(nil).Times(3)
				mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
				mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
				mockPager.On("PageResponse").Return(armcompute.SSHPublicKeysListBySubscriptionResponse{
					SSHPublicKeysListBySubscriptionResult: armcompute.SSHPublicKeysListBySubscriptionResult{
						SSHPublicKeysGroupListResult: armcompute.SSHPublicKeysGroupListResult{
							Value: expectedResults[:2],
						},
					},
				}).Times(1)
				mockPager.On("PageResponse").Return(armcompute.SSHPublicKeysListBySubscriptionResponse{
					SSHPublicKeysListBySubscriptionResult: armcompute.SSHPublicKeysListBySubscriptionResult{
						SSHPublicKeysGroupListResult: armcompute.SSHPublicKeysGroupListResult{
							Value: expectedResults[2:],
						},
					},
				}).Times(1)

				mockCache.On("Get", "computeListAllSSHPublicKeys").Return(nil).Times(1)
				mockCache.On("Put", "computeListAllSSHPublicKeys", expectedResults).Return(false).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should hit cache and return SSH public keys",
			mocks: func(mockPager *mockSshPublicKeyListPager, mockCache *cache.MockCache) {
				mockCache.On("Get", "computeListAllSSHPublicKeys").Return(expectedResults).Times(1)
			},
			expected: expectedResults,
		},
		{
			name: "should return remote error",
			mocks: func(mockPager *mockSshPublicKeyListPager, mockCache *cache.MockCache) {
				mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
				mockPager.On("PageResponse").Return(armcompute.SSHPublicKeysListBySubscriptionResponse{
					SSHPublicKeysListBySubscriptionResult: armcompute.SSHPublicKeysListBySubscriptionResult{
						SSHPublicKeysGroupListResult: armcompute.SSHPublicKeysGroupListResult{
							Value: []*armcompute.SSHPublicKeyResource{},
						},
					},
				}).Times(1)
				mockPager.On("Err").Return(errors.New("remote error")).Times(1)

				mockCache.On("Get", "computeListAllSSHPublicKeys").Return(nil).Times(1)
			},
			wantErr: "remote error",
		},
		{
			name: "should return remote error after fetching all pages",
			mocks: func(mockPager *mockSshPublicKeyListPager, mockCache *cache.MockCache) {
				mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
				mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
				mockPager.On("PageResponse").Return(armcompute.SSHPublicKeysListBySubscriptionResponse{
					SSHPublicKeysListBySubscriptionResult: armcompute.SSHPublicKeysListBySubscriptionResult{
						SSHPublicKeysGroupListResult: armcompute.SSHPublicKeysGroupListResult{
							Value: []*armcompute.SSHPublicKeyResource{},
						},
					},
				}).Times(1)
				mockPager.On("Err").Return(nil).Times(1)
				mockPager.On("Err").Return(errors.New("remote error")).Times(1)

				mockCache.On("Get", "computeListAllSSHPublicKeys").Return(nil).Times(1)
			},
			wantErr: "remote error",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := &mockSshPublicKeyClient{}
			mockPager := &mockSshPublicKeyListPager{}
			mockCache := &cache.MockCache{}

			fakeClient.On("ListBySubscription", mock.Anything).Maybe().Return(mockPager)

			tt.mocks(mockPager, mockCache)

			s := &computeRepository{
				sshPublicKeyClient: fakeClient,
				cache:              mockCache,
			}
			got, err := s.ListAllSSHPublicKeys()
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

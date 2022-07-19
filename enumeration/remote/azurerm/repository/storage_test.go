package repository

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_ListAllStorageAccount_MultiplesResults(t *testing.T) {

	expected := []*armstorage.StorageAccount{
		{
			TrackedResource: armstorage.TrackedResource{
				Resource: armstorage.Resource{
					ID: func(s string) *string { return &s }("account1"),
				},
			},
		},
		{
			TrackedResource: armstorage.TrackedResource{
				Resource: armstorage.Resource{
					ID: func(s string) *string { return &s }("account2"),
				},
			},
		},
		{
			TrackedResource: armstorage.TrackedResource{
				Resource: armstorage.Resource{
					ID: func(s string) *string { return &s }("account3"),
				},
			},
		},
		{
			TrackedResource: armstorage.TrackedResource{
				Resource: armstorage.Resource{
					ID: func(s string) *string { return &s }("account4"),
				},
			},
		},
	}

	fakeClient := &mockStorageAccountClient{}

	mockPager := &mockStorageAccountListPager{}
	mockPager.On("Err").Return(nil).Times(3)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
	mockPager.On("PageResponse").Return(armstorage.StorageAccountsListResponse{
		StorageAccountsListResult: armstorage.StorageAccountsListResult{
			StorageAccountListResult: armstorage.StorageAccountListResult{
				Value: []*armstorage.StorageAccount{
					{
						TrackedResource: armstorage.TrackedResource{
							Resource: armstorage.Resource{
								ID: func(s string) *string { return &s }("account1"),
							},
						},
					},
					{
						TrackedResource: armstorage.TrackedResource{
							Resource: armstorage.Resource{
								ID: func(s string) *string { return &s }("account2"),
							},
						},
					},
				},
			},
		},
	}).Times(1)
	mockPager.On("PageResponse").Return(armstorage.StorageAccountsListResponse{
		StorageAccountsListResult: armstorage.StorageAccountsListResult{
			StorageAccountListResult: armstorage.StorageAccountListResult{
				Value: []*armstorage.StorageAccount{
					{
						TrackedResource: armstorage.TrackedResource{
							Resource: armstorage.Resource{
								ID: func(s string) *string { return &s }("account3"),
							},
						},
					},
					{
						TrackedResource: armstorage.TrackedResource{
							Resource: armstorage.Resource{
								ID: func(s string) *string { return &s }("account4"),
							},
						},
					},
				},
			},
		},
	}).Times(1)

	fakeClient.On("List", mock.Anything).Return(mockPager)

	c := &cache.MockCache{}
	c.On("GetAndLock", "ListAllStorageAccount").Return(nil).Times(1)
	c.On("Unlock", "ListAllStorageAccount").Times(1)
	c.On("Put", "ListAllStorageAccount", expected).Return(true).Times(1)
	s := &storageRepository{
		storageAccountsClient: fakeClient,
		cache:                 c,
	}
	got, err := s.ListAllStorageAccount()
	if err != nil {
		t.Errorf("ListAllStorageAccount() error = %v", err)
		return
	}

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllStorageAccount() got = %v, want %v", got, expected)
	}
}

func Test_ListAllStorageAccount_MultiplesResults_WithCache(t *testing.T) {

	expected := []*armstorage.StorageAccount{
		{
			TrackedResource: armstorage.TrackedResource{
				Resource: armstorage.Resource{
					ID: func(s string) *string { return &s }("account1"),
				},
			},
		},
	}

	fakeClient := &mockStorageAccountClient{}

	c := &cache.MockCache{}
	c.On("GetAndLock", "ListAllStorageAccount").Return(expected).Times(1)
	c.On("Unlock", "ListAllStorageAccount").Times(1)
	s := &storageRepository{
		storageAccountsClient: fakeClient,
		cache:                 c,
	}
	got, err := s.ListAllStorageAccount()
	if err != nil {
		t.Errorf("ListAllStorageAccount() error = %v", err)
		return
	}

	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("ListAllStorageAccount() got = %v, want %v", got, expected)
	}
}

func Test_ListAllStorageAccount_Error(t *testing.T) {

	fakeClient := &mockStorageAccountClient{}

	expectedErr := errors.New("unexpected error")

	mockPager := &mockStorageAccountListPager{}
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
	mockPager.On("PageResponse").Return(armstorage.StorageAccountsListResponse{}).Times(1)

	fakeClient.On("List", mock.Anything).Return(mockPager)

	s := &storageRepository{
		storageAccountsClient: fakeClient,
		cache:                 cache.New(0),
	}
	got, err := s.ListAllStorageAccount()

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, got)
}

func Test_ListAllStorageContainer_MultiplesResults(t *testing.T) {

	account := armstorage.StorageAccount{
		Properties: &armstorage.StorageAccountProperties{
			PrimaryEndpoints: &armstorage.Endpoints{
				Blob: to.StringPtr("https://testeliedriftctl.blob.core.windows.net/"),
			},
		},
		TrackedResource: armstorage.TrackedResource{
			Resource: armstorage.Resource{
				ID:   to.StringPtr("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/foobar/providers/Microsoft.Storage/storageAccounts/testeliedriftctl"),
				Name: to.StringPtr("testeliedriftctl"),
			},
		},
	}

	expected := []string{
		"https://testeliedriftctl.blob.core.windows.net/container1",
		"https://testeliedriftctl.blob.core.windows.net/container2",
		"https://testeliedriftctl.blob.core.windows.net/container3",
		"https://testeliedriftctl.blob.core.windows.net/container4",
	}

	fakeClient := &mockBlobContainerClient{}

	mockPager := &mockBlobContainerListPager{}
	mockPager.On("Err").Return(nil).Times(3)
	mockPager.On("NextPage", mock.Anything).Return(true).Times(2)
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
	mockPager.On("PageResponse").Return(armstorage.BlobContainersListResponse{
		BlobContainersListResult: armstorage.BlobContainersListResult{
			ListContainerItems: armstorage.ListContainerItems{
				Value: []*armstorage.ListContainerItem{
					{
						AzureEntityResource: armstorage.AzureEntityResource{
							Resource: armstorage.Resource{
								Name: to.StringPtr("container1"),
							},
						},
					},
					{
						AzureEntityResource: armstorage.AzureEntityResource{
							Resource: armstorage.Resource{
								Name: to.StringPtr("container2"),
							},
						},
					},
				},
			},
		},
	}).Times(1)
	mockPager.On("PageResponse").Return(armstorage.BlobContainersListResponse{
		BlobContainersListResult: armstorage.BlobContainersListResult{
			ListContainerItems: armstorage.ListContainerItems{
				Value: []*armstorage.ListContainerItem{
					{
						AzureEntityResource: armstorage.AzureEntityResource{
							Resource: armstorage.Resource{
								Name: to.StringPtr("container3"),
							},
						},
					},
					{
						AzureEntityResource: armstorage.AzureEntityResource{
							Resource: armstorage.Resource{
								Name: to.StringPtr("container4"),
							},
						},
					},
				},
			},
		},
	}).Times(1)

	fakeClient.On("List", "foobar", "testeliedriftctl", (*armstorage.BlobContainersListOptions)(nil)).Return(mockPager)

	c := &cache.MockCache{}
	c.On("Get", "ListAllStorageContainer_testeliedriftctl").Return(nil).Times(1)
	c.On("Put", "ListAllStorageContainer_testeliedriftctl", expected).Return(true).Times(1)
	s := &storageRepository{
		blobContainerClient: fakeClient,
		cache:               c,
	}
	got, err := s.ListAllStorageContainer(&account)
	if err != nil {
		t.Errorf("ListAllStorageAccount() error = %v", err)
		return
	}

	mockPager.AssertExpectations(t)
	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	assert.Equal(t, expected, got)
}

func Test_ListAllStorageContainer_MultiplesResults_WithCache(t *testing.T) {

	account := armstorage.StorageAccount{
		TrackedResource: armstorage.TrackedResource{
			Resource: armstorage.Resource{
				ID:   to.StringPtr("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/foobar/providers/Microsoft.Storage/storageAccounts/testeliedriftctl"),
				Name: to.StringPtr("testeliedriftctl"),
			},
		},
	}

	expected := []string{
		"https://testeliedriftctl.blob.core.windows.net/container1",
	}

	fakeClient := &mockBlobContainerClient{}

	c := &cache.MockCache{}
	c.On("Get", "ListAllStorageContainer_testeliedriftctl").Return(expected).Times(1)
	s := &storageRepository{
		blobContainerClient: fakeClient,
		cache:               c,
	}
	got, err := s.ListAllStorageContainer(&account)
	if err != nil {
		t.Errorf("ListAllStorageAccount() error = %v", err)
		return
	}

	fakeClient.AssertExpectations(t)
	c.AssertExpectations(t)

	assert.Equal(t, expected, got)
}

func Test_ListAllStorageContainer_InvalidStorageAccountResourceID(t *testing.T) {

	account := armstorage.StorageAccount{
		TrackedResource: armstorage.TrackedResource{
			Resource: armstorage.Resource{
				ID:   to.StringPtr("foobar"),
				Name: to.StringPtr(""),
			},
		},
	}

	fakeClient := &mockBlobContainerClient{}

	s := &storageRepository{
		blobContainerClient: fakeClient,
		cache:               cache.New(0),
	}
	got, err := s.ListAllStorageContainer(&account)

	fakeClient.AssertExpectations(t)

	assert.Nil(t, got)
	assert.Equal(t, "parsing failed for foobar. Invalid resource Id format", err.Error())
}

func Test_ListAllStorageContainer_Error(t *testing.T) {

	account := armstorage.StorageAccount{
		TrackedResource: armstorage.TrackedResource{
			Resource: armstorage.Resource{
				ID:   to.StringPtr("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/foobar/providers/Microsoft.Storage/storageAccounts/testeliedriftctl"),
				Name: to.StringPtr("testeliedriftctl"),
			},
		},
	}

	expectedErr := errors.New("sample error")

	fakeClient := &mockBlobContainerClient{}
	mockPager := &mockBlobContainerListPager{}
	mockPager.On("NextPage", mock.Anything).Return(true).Times(1)
	mockPager.On("Err").Return(expectedErr).Times(1)
	mockPager.On("PageResponse").Return(armstorage.BlobContainersListResponse{}).Times(1)

	fakeClient.On("List", "foobar", "testeliedriftctl", (*armstorage.BlobContainersListOptions)(nil)).Return(mockPager)

	s := &storageRepository{
		blobContainerClient: fakeClient,
		cache:               cache.New(0),
	}
	got, err := s.ListAllStorageContainer(&account)

	fakeClient.AssertExpectations(t)
	mockPager.AssertExpectations(t)

	assert.Nil(t, got)
	assert.Equal(t, expectedErr, err)
}

func Test_ListAllStorageContainer_IgnoredError(t *testing.T) {

	account := armstorage.StorageAccount{
		TrackedResource: armstorage.TrackedResource{
			Resource: armstorage.Resource{
				ID:   to.StringPtr("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/foobar/providers/Microsoft.Storage/storageAccounts/testeliedriftctl"),
				Name: to.StringPtr("testeliedriftctl"),
			},
		},
	}

	fakeClient := &mockBlobContainerClient{}
	mockPager := &mockBlobContainerListPager{}
	mockPager.On("NextPage", mock.Anything).Return(false).Times(1)
	mockPager.On("Err").Return(runtime.NewResponseError(
		errors.New("{\"error\":{\"code\":\"FeatureNotSupportedForAccount\",\"message\":\"Blob is not supported for the account.\"}}"),
		nil),
	).Times(1)

	fakeClient.On("List", "foobar", "testeliedriftctl", (*armstorage.BlobContainersListOptions)(nil)).Return(mockPager)

	s := &storageRepository{
		blobContainerClient: fakeClient,
		cache:               cache.New(0),
	}
	got, err := s.ListAllStorageContainer(&account)

	fakeClient.AssertExpectations(t)
	mockPager.AssertExpectations(t)

	assert.Empty(t, got)
	assert.Equal(t, nil, err)
}

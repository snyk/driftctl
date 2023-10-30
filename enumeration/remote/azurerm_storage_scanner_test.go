package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/azurerm"
	"github.com/snyk/driftctl/enumeration/remote/azurerm/repository"
	"github.com/snyk/driftctl/enumeration/remote/common"
	error2 "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceazure "github.com/snyk/driftctl/enumeration/resource/azurerm"
	"github.com/snyk/driftctl/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAzurermStorageAccount(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockStorageRespository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no storage accounts",
			mocks: func(repository *repository.MockStorageRespository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllStorageAccount").Return([]*armstorage.StorageAccount{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing storage accounts",
			mocks: func(repository *repository.MockStorageRespository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllStorageAccount").Return(nil, dummyError)
			},
			wantErr: error2.NewResourceListingError(dummyError, resourceazure.AzureStorageAccountResourceType),
		},
		{
			test: "multiple storage accounts",
			mocks: func(repository *repository.MockStorageRespository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllStorageAccount").Return([]*armstorage.StorageAccount{
					{
						TrackedResource: armstorage.TrackedResource{
							Resource: armstorage.Resource{
								ID: func(s string) *string { return &s }("testeliedriftctl1"),
							},
						},
					},
					{
						TrackedResource: armstorage.TrackedResource{
							Resource: armstorage.Resource{
								ID: func(s string) *string { return &s }("testeliedriftctl2"),
							},
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "testeliedriftctl1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzureStorageAccountResourceType)

				assert.Equal(t, got[1].ResourceId(), "testeliedriftctl2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzureStorageAccountResourceType)
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockStorageRespository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.StorageRespository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermStorageAccountEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestAzurermStorageContainer(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockStorageRespository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no storage accounts",
			mocks: func(repository *repository.MockStorageRespository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllStorageAccount").Return([]*armstorage.StorageAccount{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "no storage containers",
			mocks: func(repository *repository.MockStorageRespository, alerter *mocks.AlerterInterface) {
				account1 := &armstorage.StorageAccount{
					TrackedResource: armstorage.TrackedResource{
						Resource: armstorage.Resource{
							ID: func(s string) *string { return &s }("testeliedriftctl1"),
						},
					},
				}
				account2 := &armstorage.StorageAccount{
					TrackedResource: armstorage.TrackedResource{
						Resource: armstorage.Resource{
							ID: func(s string) *string { return &s }("testeliedriftctl1"),
						},
					},
				}
				repository.On("ListAllStorageAccount").Return([]*armstorage.StorageAccount{
					account1,
					account2,
				}, nil)
				repository.On("ListAllStorageContainer", account1).Return([]string{}, nil)
				repository.On("ListAllStorageContainer", account2).Return([]string{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing storage accounts",
			mocks: func(repository *repository.MockStorageRespository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllStorageAccount").Return(nil, dummyError)
			},
			wantErr: error2.NewResourceListingErrorWithType(dummyError, resourceazure.AzureStorageContainerResourceType, resourceazure.AzureStorageAccountResourceType),
		},
		{
			test: "error listing storage container",
			mocks: func(repository *repository.MockStorageRespository, alerter *mocks.AlerterInterface) {
				account := &armstorage.StorageAccount{
					TrackedResource: armstorage.TrackedResource{
						Resource: armstorage.Resource{
							ID: func(s string) *string { return &s }("testeliedriftctl1"),
						},
					},
				}
				repository.On("ListAllStorageAccount").Return([]*armstorage.StorageAccount{account}, nil)
				repository.On("ListAllStorageContainer", account).Return(nil, dummyError)
			},
			wantErr: error2.NewResourceListingError(dummyError, resourceazure.AzureStorageContainerResourceType),
		},
		{
			test: "multiple storage containers",
			mocks: func(repository *repository.MockStorageRespository, alerter *mocks.AlerterInterface) {
				account1 := &armstorage.StorageAccount{
					TrackedResource: armstorage.TrackedResource{
						Resource: armstorage.Resource{
							ID: func(s string) *string { return &s }("testeliedriftctl1"),
						},
					},
				}
				account2 := &armstorage.StorageAccount{
					TrackedResource: armstorage.TrackedResource{
						Resource: armstorage.Resource{
							ID: func(s string) *string { return &s }("testeliedriftctl2"),
						},
					},
				}
				repository.On("ListAllStorageAccount").Return([]*armstorage.StorageAccount{
					account1,
					account2,
				}, nil)
				repository.On("ListAllStorageContainer", account1).Return([]string{"https://testeliedriftctl1.blob.core.windows.net/container1", "https://testeliedriftctl1.blob.core.windows.net/container2"}, nil)
				repository.On("ListAllStorageContainer", account2).Return([]string{"https://testeliedriftctl2.blob.core.windows.net/container3", "https://testeliedriftctl2.blob.core.windows.net/container4"}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 4)

				for _, container := range got {
					assert.Equal(t, container.ResourceType(), resourceazure.AzureStorageContainerResourceType)
				}

				assert.Equal(t, got[0].ResourceId(), "https://testeliedriftctl1.blob.core.windows.net/container1")
				assert.Equal(t, got[1].ResourceId(), "https://testeliedriftctl1.blob.core.windows.net/container2")
				assert.Equal(t, got[2].ResourceId(), "https://testeliedriftctl2.blob.core.windows.net/container3")
				assert.Equal(t, got[3].ResourceId(), "https://testeliedriftctl2.blob.core.windows.net/container4")
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockStorageRespository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.StorageRespository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermStorageContainerEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

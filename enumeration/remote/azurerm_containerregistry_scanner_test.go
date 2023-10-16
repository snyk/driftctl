package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/azurerm"
	"github.com/snyk/driftctl/enumeration/remote/azurerm/repository"
	"github.com/snyk/driftctl/enumeration/remote/common"
	error2 "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceazure "github.com/snyk/driftctl/enumeration/resource/azurerm"
	"github.com/snyk/driftctl/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAzurermContainerRegistry(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockContainerRegistryRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no container registry",
			mocks: func(repository *repository.MockContainerRegistryRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllContainerRegistries").Return([]*armcontainerregistry.Registry{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing container registry",
			mocks: func(repository *repository.MockContainerRegistryRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllContainerRegistries").Return(nil, dummyError)
			},
			wantErr: error2.NewResourceListingError(dummyError, resourceazure.AzureContainerRegistryResourceType),
		},
		{
			test: "multiple container registries",
			mocks: func(repository *repository.MockContainerRegistryRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllContainerRegistries").Return([]*armcontainerregistry.Registry{
					{
						Resource: armcontainerregistry.Resource{
							ID: to.StringPtr("registry1"),
						},
					},
					{
						Resource: armcontainerregistry.Resource{
							ID: to.StringPtr("registry2"),
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "registry1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzureContainerRegistryResourceType)

				assert.Equal(t, got[1].ResourceId(), "registry2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzureContainerRegistryResourceType)
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockContainerRegistryRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ContainerRegistryRepository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermContainerRegistryEnumerator(repo, factory))

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

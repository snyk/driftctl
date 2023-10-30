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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceazure "github.com/snyk/driftctl/enumeration/resource/azurerm"
	"github.com/snyk/driftctl/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAzurermResourceGroups(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockResourcesRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no resource group",
			mocks: func(repository *repository.MockResourcesRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllResourceGroups").Return([]*armresources.ResourceGroup{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing resource groups",
			mocks: func(repository *repository.MockResourcesRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllResourceGroups").Return(nil, dummyError)
			},
			wantErr: error2.NewResourceListingError(dummyError, resourceazure.AzureResourceGroupResourceType),
		},
		{
			test: "multiple resource groups",
			mocks: func(repository *repository.MockResourcesRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllResourceGroups").Return([]*armresources.ResourceGroup{
					{
						ID:   to.StringPtr("group1"),
						Name: to.StringPtr("group1"),
					},
					{
						ID:   to.StringPtr("group2"),
						Name: to.StringPtr("group2"),
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "group1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzureResourceGroupResourceType)

				assert.Equal(t, got[1].ResourceId(), "group2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzureResourceGroupResourceType)
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockResourcesRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ResourcesRepository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermResourceGroupEnumerator(repo, factory))

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

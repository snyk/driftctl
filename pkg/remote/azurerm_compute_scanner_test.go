package remote

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	remoteerr "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceazure "github.com/cloudskiff/driftctl/pkg/resource/azurerm"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAzurermCompute_Image(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockComputeRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no images",
			mocks: func(repository *repository.MockComputeRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllImages").Return([]*armcompute.Image{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing images",
			mocks: func(repository *repository.MockComputeRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllImages").Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceazure.AzureImageResourceType),
		},
		{
			test: "multiple images including an invalid ID",
			mocks: func(repository *repository.MockComputeRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllImages").Return([]*armcompute.Image{
					{
						Resource: armcompute.Resource{
							ID:   to.StringPtr("/subscriptions/4e411884-65b0-4911-bc80-52f9a21942a2/resourceGroups/testgroup/providers/Microsoft.Compute/images/image1"),
							Name: to.StringPtr("image1"),
						},
					},
					{
						Resource: armcompute.Resource{
							ID:   to.StringPtr("/subscriptions/4e411884-65b0-4911-bc80-52f9a21942a2/resourceGroups/testgroup/providers/Microsoft.Compute/images/image2"),
							Name: to.StringPtr("image2"),
						},
					},
					{
						Resource: armcompute.Resource{
							ID:   to.StringPtr("/invalid-id/image3"),
							Name: to.StringPtr("image3"),
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "/subscriptions/4e411884-65b0-4911-bc80-52f9a21942a2/resourceGroups/testgroup/providers/Microsoft.Compute/images/image1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzureImageResourceType)

				assert.Equal(t, got[1].ResourceId(), "/subscriptions/4e411884-65b0-4911-bc80-52f9a21942a2/resourceGroups/testgroup/providers/Microsoft.Compute/images/image2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzureImageResourceType)
			},
		},
	}

	providerVersion := "2.71.0"
	schemaRepository := testresource.InitFakeSchemaRepository("azurerm", providerVersion)
	resourceazure.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			scanOptions := ScannerOptions{}
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockComputeRepository{}
			c.mocks(fakeRepo, alerter)

			remoteLibrary.AddEnumerator(azurerm.NewAzurermImageEnumerator(fakeRepo, factory))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
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

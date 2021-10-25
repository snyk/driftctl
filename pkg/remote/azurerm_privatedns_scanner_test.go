package remote

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	error2 "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceazure "github.com/cloudskiff/driftctl/pkg/resource/azurerm"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAzurermPrivateDNSZone(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockPrivateDNSRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no PrivateZone",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing PrivateZone",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return(nil, dummyError)
			},
			wantErr: error2.NewResourceListingError(dummyError, resourceazure.AzurePrivateDNSZoneResourceType),
		},
		{
			test: "multiple PrivateZone",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID: to.StringPtr("driftctlzone1"),
							},
						},
					},
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID: to.StringPtr("driftctlzone2"),
							},
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "driftctlzone1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzurePrivateDNSZoneResourceType)

				assert.Equal(t, got[1].ResourceId(), "driftctlzone2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzurePrivateDNSZoneResourceType)
			},
		},
	}

	providerVersion := "2.71.0"
	schemaRepository := testresource.InitFakeSchemaRepository("azurerm", providerVersion)
	resourceazure.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {

			scanOptions := ScannerOptions{Deep: true}
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockPrivateDNSRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.PrivateDNSRepository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermPrivateDNSZoneEnumerator(repo, factory))

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

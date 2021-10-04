package remote

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/network/armnetwork"
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

func TestAzurermVirtualNetwork(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockNetworkRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no virtual network",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVirtualNetworks").Return([]*armnetwork.VirtualNetwork{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing virtual network",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVirtualNetworks").Return(nil, dummyError)
			},
			wantErr: error2.NewResourceListingError(dummyError, resourceazure.AzureVirtualNetworkResourceType),
		},
		{
			test: "multiple virtual network",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVirtualNetworks").Return([]*armnetwork.VirtualNetwork{
					{
						Resource: armnetwork.Resource{
							ID:   to.StringPtr("network1"),
							Name: to.StringPtr("network1"),
						},
					},
					{
						Resource: armnetwork.Resource{
							ID:   to.StringPtr("network2"),
							Name: to.StringPtr("network2"),
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "network1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzureVirtualNetworkResourceType)

				assert.Equal(t, got[1].ResourceId(), "network2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzureVirtualNetworkResourceType)
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
			fakeRepo := &repository.MockNetworkRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.NetworkRepository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermVirtualNetworkEnumerator(repo, factory))

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

func TestAzurermRouteTables(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockNetworkRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no route tables",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRouteTables").Return([]*armnetwork.RouteTable{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing route tables",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRouteTables").Return(nil, dummyError)
			},
			wantErr: error2.NewResourceListingError(dummyError, resourceazure.AzureRouteTableResourceType),
		},
		{
			test: "multiple route tables",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRouteTables").Return([]*armnetwork.RouteTable{
					{
						Resource: armnetwork.Resource{
							ID:   to.StringPtr("route1"),
							Name: to.StringPtr("route1"),
						},
					},
					{
						Resource: armnetwork.Resource{
							ID:   to.StringPtr("route2"),
							Name: to.StringPtr("route2"),
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "route1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzureRouteTableResourceType)

				assert.Equal(t, got[1].ResourceId(), "route2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzureRouteTableResourceType)
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
			fakeRepo := &repository.MockNetworkRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.NetworkRepository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermRouteTableEnumerator(repo, factory))

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

package remote

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	remoteerr "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceazure "github.com/cloudskiff/driftctl/pkg/resource/azurerm"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraformtest "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAzurermPrivateDNSZone(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockPrivateDNSRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no private zone",
			dirName: "azurerm_private_dns_private_zone_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{}, nil)
			},
		},
		{
			test:    "error listing private zones",
			dirName: "azurerm_private_dns_private_zone_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceazure.AzurePrivateDNSZoneResourceType),
		},
		{
			test:    "multiple private zones",
			dirName: "azurerm_private_dns_private_zone_multiple",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com"),
								Name: to.StringPtr("thisisatestusingtf.com"),
							},
						},
					},
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf2.com"),
								Name: to.StringPtr("thisisatestusingtf2.com"),
							},
						},
					},
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/testmartin.com"),
								Name: to.StringPtr("testmartin.com"),
							},
						},
					},
				}, nil)
			},
		},
	}

	providerVersion := "2.71.0"
	schemaRepository := testresource.InitFakeSchemaRepository("azurerm", providerVersion)
	resourceazure.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockPrivateDNSRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.PrivateDNSRepository = fakeRepo
			providerVersion := "2.71.0"
			realProvider, err := terraformtest.InitTestAzureProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraformtest.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{})
				if err != nil {
					t.Fatal(err)
				}
				con := arm.NewDefaultConnection(cred, nil)
				repo = repository.NewPrivateDNSRepository(con, realProvider.GetConfig(), cache.New(0))
			}

			remoteLibrary.AddEnumerator(azurerm.NewAzurermPrivateDNSZoneEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceazure.AzurePrivateDNSZoneResourceType, common.NewGenericDetailsFetcher(resourceazure.AzurePrivateDNSZoneResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)

			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceazure.AzurePrivateDNSZoneResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestAzurermPrivateDNSARecord(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockPrivateDNSRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no private a record",
			dirName: "azurerm_private_dns_a_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{}, nil)
			},
		},
		{
			test:    "error listing private zone",
			dirName: "azurerm_private_dns_a_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceazure.AzurePrivateDNSARecordResourceType, resourceazure.AzurePrivateDNSZoneResourceType),
		},
		{
			test:    "error listing private a records",
			dirName: "azurerm_private_dns_a_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com"),
								Name: to.StringPtr("thisisatestusingtf.com"),
							},
						},
					},
				}, nil)
				repository.On("ListAllARecords", mock.Anything).Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceazure.AzurePrivateDNSARecordResourceType),
		},
		{
			test:    "multiple private a records",
			dirName: "azurerm_private_dns_a_record_multiple",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com"),
								Name: to.StringPtr("thisisatestusingtf.com"),
							},
						},
					},
				}, nil)

				repository.On("ListAllARecords", mock.Anything).Return([]*armprivatedns.RecordSet{
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/A/test"),
								Name: to.StringPtr("test"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							ARecords: []*armprivatedns.ARecord{
								{IPv4Address: to.StringPtr("10.0.180.17")},
								{IPv4Address: to.StringPtr("10.0.180.20")},
							},
						},
					},
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/A/othertest"),
								Name: to.StringPtr("othertest"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							ARecords: []*armprivatedns.ARecord{
								{IPv4Address: to.StringPtr("10.0.180.20")},
							},
						},
					},
				}, nil).Once()
			},
		},
	}

	providerVersion := "2.71.0"
	schemaRepository := testresource.InitFakeSchemaRepository("azurerm", providerVersion)
	resourceazure.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockPrivateDNSRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.PrivateDNSRepository = fakeRepo
			providerVersion := "2.71.0"
			realProvider, err := terraformtest.InitTestAzureProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraformtest.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{})
				if err != nil {
					t.Fatal(err)
				}
				con := arm.NewDefaultConnection(cred, nil)
				repo = repository.NewPrivateDNSRepository(con, realProvider.GetConfig(), cache.New(0))
			}

			remoteLibrary.AddEnumerator(azurerm.NewAzurermPrivateDNSARecordEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceazure.AzurePrivateDNSARecordResourceType, common.NewGenericDetailsFetcher(resourceazure.AzurePrivateDNSARecordResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)

			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceazure.AzurePrivateDNSARecordResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestAzurermPrivateDNSAAAARecord(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockPrivateDNSRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no private aaaa record",
			dirName: "azurerm_private_dns_aaaa_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{}, nil)
			},
		},
		{
			test:    "error listing private zone",
			dirName: "azurerm_private_dns_aaaa_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceazure.AzurePrivateDNSAAAARecordResourceType, resourceazure.AzurePrivateDNSZoneResourceType),
		},
		{
			test:    "error listing private aaaa records",
			dirName: "azurerm_private_dns_aaaa_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com"),
								Name: to.StringPtr("thisisatestusingtf.com"),
							},
						},
					},
				}, nil)
				repository.On("ListAllAAAARecords", mock.Anything).Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceazure.AzurePrivateDNSAAAARecordResourceType),
		},
		{
			test:    "multiple private aaaaa records",
			dirName: "azurerm_private_dns_aaaaa_record_multiple",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com"),
								Name: to.StringPtr("thisisatestusingtf.com"),
							},
						},
					},
				}, nil)

				repository.On("ListAllAAAARecords", mock.Anything).Return([]*armprivatedns.RecordSet{
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/AAAA/test"),
								Name: to.StringPtr("test"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							AaaaRecords: []*armprivatedns.AaaaRecord{
								{IPv6Address: to.StringPtr("fd5d:70bc:930e:d008:0000:0000:0000:7334")},
								{IPv6Address: to.StringPtr("fd5d:70bc:930e:d008::7335")},
							},
						},
					},
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/AAAA/othertest"),
								Name: to.StringPtr("othertest"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							AaaaRecords: []*armprivatedns.AaaaRecord{
								{IPv6Address: to.StringPtr("fd5d:70bc:930e:d008:0000:0000:0000:7334")},
								{IPv6Address: to.StringPtr("fd5d:70bc:930e:d008::7335")},
							},
						},
					},
				}, nil).Once()
			},
		},
	}

	providerVersion := "2.71.0"
	schemaRepository := testresource.InitFakeSchemaRepository("azurerm", providerVersion)
	resourceazure.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockPrivateDNSRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.PrivateDNSRepository = fakeRepo
			providerVersion := "2.71.0"
			realProvider, err := terraformtest.InitTestAzureProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraformtest.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{})
				if err != nil {
					t.Fatal(err)
				}
				con := arm.NewDefaultConnection(cred, nil)
				repo = repository.NewPrivateDNSRepository(con, realProvider.GetConfig(), cache.New(0))
			}

			remoteLibrary.AddEnumerator(azurerm.NewAzurermPrivateDNSAAAARecordEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceazure.AzurePrivateDNSAAAARecordResourceType, common.NewGenericDetailsFetcher(resourceazure.AzurePrivateDNSAAAARecordResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)

			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceazure.AzurePrivateDNSAAAARecordResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

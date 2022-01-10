package remote

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/mocks"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/remote/azurerm"
	"github.com/snyk/driftctl/pkg/remote/azurerm/repository"
	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/snyk/driftctl/pkg/remote/common"
	remoteerr "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	resourceazure "github.com/snyk/driftctl/pkg/resource/azurerm"
	"github.com/snyk/driftctl/pkg/terraform"
	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/goldenfile"
	testresource "github.com/snyk/driftctl/test/resource"
	terraformtest "github.com/snyk/driftctl/test/terraform"
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
				clientOptions := &arm.ClientOptions{}
				repo = repository.NewPrivateDNSRepository(cred, clientOptions, realProvider.GetConfig(), cache.New(0))
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
				clientOptions := &arm.ClientOptions{}
				repo = repository.NewPrivateDNSRepository(cred, clientOptions, realProvider.GetConfig(), cache.New(0))
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
				clientOptions := &arm.ClientOptions{}
				repo = repository.NewPrivateDNSRepository(cred, clientOptions, realProvider.GetConfig(), cache.New(0))
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

func TestAzurermPrivateDNSCNAMERecord(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockPrivateDNSRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no private cname record",
			dirName: "azurerm_private_dns_cname_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{}, nil)
			},
		},
		{
			test:    "error listing private zone",
			dirName: "azurerm_private_dns_cname_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceazure.AzurePrivateDNSCNameRecordResourceType, resourceazure.AzurePrivateDNSZoneResourceType),
		},
		{
			test:    "error listing private cname records",
			dirName: "azurerm_private_dns_cname_record_empty",
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
				repository.On("ListAllCNAMERecords", mock.Anything).Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceazure.AzurePrivateDNSCNameRecordResourceType),
		},
		{
			test:    "multiple private cname records",
			dirName: "azurerm_private_dns_cname_record_multiple",
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

				repository.On("ListAllCNAMERecords", mock.Anything).Return([]*armprivatedns.RecordSet{
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/CNAME/test"),
								Name: to.StringPtr("test"),
							},
						},
					},
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/CNAME/othertest"),
								Name: to.StringPtr("othertest"),
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
				clientOptions := &arm.ClientOptions{}
				repo = repository.NewPrivateDNSRepository(cred, clientOptions, realProvider.GetConfig(), cache.New(0))
			}

			remoteLibrary.AddEnumerator(azurerm.NewAzurermPrivateDNSCNameRecordEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceazure.AzurePrivateDNSCNameRecordResourceType, common.NewGenericDetailsFetcher(resourceazure.AzurePrivateDNSCNameRecordResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)

			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceazure.AzurePrivateDNSCNameRecordResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestAzurermPrivateDNSPTRRecord(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockPrivateDNSRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no private ptr record",
			dirName: "azurerm_private_dns_ptr_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{}, nil)
			},
		},
		{
			test:    "error listing private zone",
			dirName: "azurerm_private_dns_ptr_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceazure.AzurePrivateDNSPTRRecordResourceType, resourceazure.AzurePrivateDNSZoneResourceType),
		},
		{
			test:    "error listing private ptr records",
			dirName: "azurerm_private_dns_ptr_record_empty",
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
				repository.On("ListAllPTRRecords", mock.Anything).Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceazure.AzurePrivateDNSPTRRecordResourceType),
		},
		{
			test:    "multiple private ptra records",
			dirName: "azurerm_private_dns_ptr_record_multiple",
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

				repository.On("ListAllPTRRecords", mock.Anything).Return([]*armprivatedns.RecordSet{
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/PTR/othertestptr"),
								Name: to.StringPtr("othertestptr"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							PtrRecords: []*armprivatedns.PtrRecord{
								{Ptrdname: to.StringPtr("ptr1.thisisatestusingtf.com")},
								{Ptrdname: to.StringPtr("ptr2.thisisatestusingtf.com")},
							},
						},
					},
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/PTR/testptr"),
								Name: to.StringPtr("testptr"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							PtrRecords: []*armprivatedns.PtrRecord{
								{Ptrdname: to.StringPtr("ptr3.thisisatestusingtf.com")},
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
				clientOptions := &arm.ClientOptions{}
				repo = repository.NewPrivateDNSRepository(cred, clientOptions, realProvider.GetConfig(), cache.New(0))
			}

			remoteLibrary.AddEnumerator(azurerm.NewAzurermPrivateDNSPTRRecordEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceazure.AzurePrivateDNSPTRRecordResourceType, common.NewGenericDetailsFetcher(resourceazure.AzurePrivateDNSPTRRecordResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)

			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceazure.AzurePrivateDNSPTRRecordResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestAzurermPrivateDNSMXRecord(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockPrivateDNSRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no private mx record",
			dirName: "azurerm_private_dns_mx_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{}, nil)
			},
		},
		{
			test:    "error listing private zone",
			dirName: "azurerm_private_dns_mx_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceazure.AzurePrivateDNSMXRecordResourceType, resourceazure.AzurePrivateDNSZoneResourceType),
		},
		{
			test:    "error listing private mx records",
			dirName: "azurerm_private_dns_mx_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com"),
								Name: to.StringPtr("thisisatestusingtf.com"),
							},
						},
					},
				}, nil)
				repository.On("ListAllMXRecords", mock.Anything).Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceazure.AzurePrivateDNSMXRecordResourceType),
		},
		{
			test:    "multiple private mx records",
			dirName: "azurerm_private_dns_mx_record_multiple",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{
					{
						TrackedResource: armprivatedns.TrackedResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com"),
								Name: to.StringPtr("thisisatestusingtf.com"),
							},
						},
					},
				}, nil)

				repository.On("ListAllMXRecords", mock.Anything).Return([]*armprivatedns.RecordSet{
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/MX/othertestmx"),
								Name: to.StringPtr("othertestmx"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							MxRecords: []*armprivatedns.MxRecord{
								{Exchange: to.StringPtr("ex1")},
								{Exchange: to.StringPtr("ex2")},
							},
						},
					},
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/MX/testmx"),
								Name: to.StringPtr("testmx"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							MxRecords: []*armprivatedns.MxRecord{
								{Exchange: to.StringPtr("ex1")},
								{Exchange: to.StringPtr("ex2")},
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
				clientOptions := &arm.ClientOptions{}
				repo = repository.NewPrivateDNSRepository(cred, clientOptions, realProvider.GetConfig(), cache.New(0))
			}

			remoteLibrary.AddEnumerator(azurerm.NewAzurermPrivateDNSMXRecordEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceazure.AzurePrivateDNSMXRecordResourceType, common.NewGenericDetailsFetcher(resourceazure.AzurePrivateDNSMXRecordResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)

			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceazure.AzurePrivateDNSMXRecordResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestAzurermPrivateDNSSRVRecord(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockPrivateDNSRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no private srv record",
			dirName: "azurerm_private_dns_srv_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{}, nil)
			},
		},
		{
			test:    "error listing private zone",
			dirName: "azurerm_private_dns_srv_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceazure.AzurePrivateDNSSRVRecordResourceType, resourceazure.AzurePrivateDNSZoneResourceType),
		},
		{
			test:    "error listing private srv records",
			dirName: "azurerm_private_dns_srv_record_empty",
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
				repository.On("ListAllSRVRecords", mock.Anything).Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceazure.AzurePrivateDNSSRVRecordResourceType),
		},
		{
			test:    "multiple private srv records",
			dirName: "azurerm_private_dns_srv_record_multiple",
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

				repository.On("ListAllSRVRecords", mock.Anything).Return([]*armprivatedns.RecordSet{
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/SRV/othertestptr"),
								Name: to.StringPtr("othertestptr"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							SrvRecords: []*armprivatedns.SrvRecord{
								{Target: to.StringPtr("srv1.thisisatestusingtf.com")},
								{Target: to.StringPtr("srv2.thisisatestusingtf.com")},
							},
						},
					},
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/SRV/testptr"),
								Name: to.StringPtr("testptr"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							PtrRecords: []*armprivatedns.PtrRecord{
								{Ptrdname: to.StringPtr("srv3.thisisatestusingtf.com")},
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
				clientOptions := &arm.ClientOptions{}
				repo = repository.NewPrivateDNSRepository(cred, clientOptions, realProvider.GetConfig(), cache.New(0))
			}

			remoteLibrary.AddEnumerator(azurerm.NewAzurermPrivateDNSSRVRecordEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceazure.AzurePrivateDNSSRVRecordResourceType, common.NewGenericDetailsFetcher(resourceazure.AzurePrivateDNSSRVRecordResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)

			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceazure.AzurePrivateDNSSRVRecordResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestAzurermPrivateDNSTXTRecord(t *testing.T) {

	dummyError := errors.New("this is an error")

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockPrivateDNSRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no private txt record",
			dirName: "azurerm_private_dns_txt_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return([]*armprivatedns.PrivateZone{}, nil)
			},
		},
		{
			test:    "error listing private zone",
			dirName: "azurerm_private_dns_txt_record_empty",
			mocks: func(repository *repository.MockPrivateDNSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPrivateZones").Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceazure.AzurePrivateDNSTXTRecordResourceType, resourceazure.AzurePrivateDNSZoneResourceType),
		},
		{
			test:    "error listing private txt records",
			dirName: "azurerm_private_dns_txt_record_empty",
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
				repository.On("ListAllTXTRecords", mock.Anything).Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingError(dummyError, resourceazure.AzurePrivateDNSTXTRecordResourceType),
		},
		{
			test:    "multiple private txt records",
			dirName: "azurerm_private_dns_txt_record_multiple",
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

				repository.On("ListAllTXTRecords", mock.Anything).Return([]*armprivatedns.RecordSet{
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/TXT/othertesttxt"),
								Name: to.StringPtr("othertesttxt"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							TxtRecords: []*armprivatedns.TxtRecord{
								{Value: []*string{to.StringPtr("this is value line 1")}},
								{Value: []*string{to.StringPtr("this is value line 2")}},
							},
						},
					},
					{
						ProxyResource: armprivatedns.ProxyResource{
							Resource: armprivatedns.Resource{
								ID:   to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/martin-dev/providers/Microsoft.Network/privateDnsZones/thisisatestusingtf.com/TXT/testtxt"),
								Name: to.StringPtr("testtxt"),
							},
						},
						Properties: &armprivatedns.RecordSetProperties{
							PtrRecords: []*armprivatedns.PtrRecord{
								{Ptrdname: to.StringPtr("this is value line 3")},
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
				clientOptions := &arm.ClientOptions{}
				repo = repository.NewPrivateDNSRepository(cred, clientOptions, realProvider.GetConfig(), cache.New(0))
			}

			remoteLibrary.AddEnumerator(azurerm.NewAzurermPrivateDNSTXTRecordEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceazure.AzurePrivateDNSTXTRecordResourceType, common.NewGenericDetailsFetcher(resourceazure.AzurePrivateDNSTXTRecordResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)

			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceazure.AzurePrivateDNSTXTRecordResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

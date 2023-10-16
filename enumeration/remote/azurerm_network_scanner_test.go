package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/azurerm"
	"github.com/snyk/driftctl/enumeration/remote/azurerm/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	error2 "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceazure "github.com/snyk/driftctl/enumeration/resource/azurerm"
	"github.com/snyk/driftctl/mocks"

	"github.com/snyk/driftctl/test/goldenfile"
	terraform2 "github.com/snyk/driftctl/test/terraform"
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

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockNetworkRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.NetworkRepository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermVirtualNetworkEnumerator(repo, factory))

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

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockNetworkRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.NetworkRepository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermRouteTableEnumerator(repo, factory))

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

func TestAzurermRoutes(t *testing.T) {
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
			test: "no routes",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRouteTables").Return([]*armnetwork.RouteTable{
					{
						Properties: &armnetwork.RouteTablePropertiesFormat{
							Routes: []*armnetwork.Route{},
						},
					},
					{
						Properties: &armnetwork.RouteTablePropertiesFormat{
							Routes: []*armnetwork.Route{},
						},
					},
				}, nil)
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
			wantErr: error2.NewResourceListingErrorWithType(dummyError, resourceazure.AzureRouteResourceType, resourceazure.AzureRouteTableResourceType),
		},
		{
			test: "multiple routes",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllRouteTables").Return([]*armnetwork.RouteTable{
					{
						Resource: armnetwork.Resource{
							Name: to.StringPtr("table1"),
						},
						Properties: &armnetwork.RouteTablePropertiesFormat{
							Routes: []*armnetwork.Route{
								{
									SubResource: armnetwork.SubResource{
										ID: to.StringPtr("route1"),
									},
									Name: to.StringPtr("route1"),
								},
								{
									SubResource: armnetwork.SubResource{
										ID: to.StringPtr("route2"),
									},
									Name: to.StringPtr("route2"),
								},
							},
						},
					},
					{
						Resource: armnetwork.Resource{
							Name: to.StringPtr("table2"),
						},
						Properties: &armnetwork.RouteTablePropertiesFormat{
							Routes: []*armnetwork.Route{
								{
									SubResource: armnetwork.SubResource{
										ID: to.StringPtr("route3"),
									},
									Name: to.StringPtr("route3"),
								},
								{
									SubResource: armnetwork.SubResource{
										ID: to.StringPtr("route4"),
									},
									Name: to.StringPtr("route4"),
								},
							},
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 4)

				assert.Equal(t, "route1", got[0].ResourceId())
				assert.Equal(t, resourceazure.AzureRouteResourceType, got[0].ResourceType())

				assert.Equal(t, "route2", got[1].ResourceId())
				assert.Equal(t, resourceazure.AzureRouteResourceType, got[1].ResourceType())

				assert.Equal(t, "route3", got[2].ResourceId())
				assert.Equal(t, resourceazure.AzureRouteResourceType, got[2].ResourceType())

				assert.Equal(t, "route4", got[3].ResourceId())
				assert.Equal(t, resourceazure.AzureRouteResourceType, got[3].ResourceType())
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockNetworkRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.NetworkRepository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermRouteEnumerator(repo, factory))

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

func TestAzurermSubnets(t *testing.T) {
	dummyError := errors.New("this is an error")

	networks := []*armnetwork.VirtualNetwork{
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("network1"),
			},
		},
		{
			Resource: armnetwork.Resource{
				ID: to.StringPtr("network2"),
			},
		},
	}

	tests := []struct {
		test           string
		mocks          func(*repository.MockNetworkRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no subnets",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVirtualNetworks").Return(networks, nil)
				repository.On("ListAllSubnets", networks[0]).Return([]*armnetwork.Subnet{}, nil).Times(1)
				repository.On("ListAllSubnets", networks[1]).Return([]*armnetwork.Subnet{}, nil).Times(1)
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
			wantErr: error2.NewResourceListingErrorWithType(dummyError, resourceazure.AzureSubnetResourceType, resourceazure.AzureVirtualNetworkResourceType),
		},
		{
			test: "error listing subnets",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVirtualNetworks").Return(networks, nil)
				repository.On("ListAllSubnets", networks[0]).Return(nil, dummyError).Times(1)
			},
			wantErr: error2.NewResourceListingError(dummyError, resourceazure.AzureSubnetResourceType),
		},
		{
			test: "multiple subnets",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllVirtualNetworks").Return(networks, nil)
				repository.On("ListAllSubnets", networks[0]).Return([]*armnetwork.Subnet{
					{
						SubResource: armnetwork.SubResource{
							ID: to.StringPtr("subnet1"),
						},
					},
					{
						SubResource: armnetwork.SubResource{
							ID: to.StringPtr("subnet2"),
						},
					},
				}, nil).Times(1)
				repository.On("ListAllSubnets", networks[1]).Return([]*armnetwork.Subnet{
					{
						SubResource: armnetwork.SubResource{
							ID: to.StringPtr("subnet3"),
						},
					},
					{
						SubResource: armnetwork.SubResource{
							ID: to.StringPtr("subnet4"),
						},
					},
				}, nil).Times(1)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 4)

				assert.Equal(t, got[0].ResourceId(), "subnet1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzureSubnetResourceType)

				assert.Equal(t, got[1].ResourceId(), "subnet2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzureSubnetResourceType)

				assert.Equal(t, got[2].ResourceId(), "subnet3")
				assert.Equal(t, got[2].ResourceType(), resourceazure.AzureSubnetResourceType)

				assert.Equal(t, got[3].ResourceId(), "subnet4")
				assert.Equal(t, got[3].ResourceType(), resourceazure.AzureSubnetResourceType)
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockNetworkRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.NetworkRepository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermSubnetEnumerator(repo, factory))

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

func TestAzurermFirewalls(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockNetworkRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no firewall",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllFirewalls").Return([]*armnetwork.AzureFirewall{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing firewalls",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllFirewalls").Return(nil, dummyError)
			},
			wantErr: error2.NewResourceListingError(dummyError, resourceazure.AzureFirewallResourceType),
		},
		{
			test: "multiple firewalls",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllFirewalls").Return([]*armnetwork.AzureFirewall{
					{
						Resource: armnetwork.Resource{
							ID:   to.StringPtr("firewall1"), // Here we don't care to have a valid ID, it is for testing purpose only
							Name: to.StringPtr("firewall1"),
						},
					},
					{
						Resource: armnetwork.Resource{
							ID:   to.StringPtr("firewall2"),
							Name: to.StringPtr("firewall2"),
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "firewall1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzureFirewallResourceType)

				assert.Equal(t, got[1].ResourceId(), "firewall2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzureFirewallResourceType)
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockNetworkRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.NetworkRepository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermFirewallsEnumerator(repo, factory))

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

func TestAzurermPublicIP(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockNetworkRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no public IP",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPublicIPAddresses").Return([]*armnetwork.PublicIPAddress{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing public IPs",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPublicIPAddresses").Return(nil, dummyError)
			},
			wantErr: error2.NewResourceListingError(dummyError, resourceazure.AzurePublicIPResourceType),
		},
		{
			test: "multiple public IP",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllPublicIPAddresses").Return([]*armnetwork.PublicIPAddress{
					{
						Resource: armnetwork.Resource{
							ID:   to.StringPtr("ip1"), // Here we don't care to have a valid ID, it is for testing purpose only
							Name: to.StringPtr("ip1"),
						},
					},
					{
						Resource: armnetwork.Resource{
							ID:   to.StringPtr("ip2"),
							Name: to.StringPtr("ip2"),
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "ip1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzurePublicIPResourceType)

				assert.Equal(t, got[1].ResourceId(), "ip2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzurePublicIPResourceType)
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockNetworkRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.NetworkRepository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermPublicIPEnumerator(repo, factory))

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

func TestAzurermSecurityGroups(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockNetworkRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			test:    "no security group",
			dirName: "azurerm_network_security_group_empty",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllSecurityGroups").Return([]*armnetwork.NetworkSecurityGroup{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test:    "error listing security groups",
			dirName: "azurerm_network_security_group_empty",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllSecurityGroups").Return(nil, dummyError)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: error2.NewResourceListingError(dummyError, resourceazure.AzureNetworkSecurityGroupResourceType),
		},
		{
			test:    "multiple security groups",
			dirName: "azurerm_network_security_group_multiple",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllSecurityGroups").Return([]*armnetwork.NetworkSecurityGroup{
					{
						Resource: armnetwork.Resource{
							ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/example-resources/providers/Microsoft.Network/networkSecurityGroups/acceptanceTestSecurityGroup1"),
							Name: to.StringPtr("acceptanceTestSecurityGroup1"),
						},
					},
					{
						Resource: armnetwork.Resource{
							ID:   to.StringPtr("/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/example-resources/providers/Microsoft.Network/networkSecurityGroups/acceptanceTestSecurityGroup2"),
							Name: to.StringPtr("acceptanceTestSecurityGroup2"),
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/example-resources/providers/Microsoft.Network/networkSecurityGroups/acceptanceTestSecurityGroup1", got[0].ResourceId())
				assert.Equal(t, resourceazure.AzureNetworkSecurityGroupResourceType, got[0].ResourceType())

				assert.Equal(t, "/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/example-resources/providers/Microsoft.Network/networkSecurityGroups/acceptanceTestSecurityGroup2", got[1].ResourceId())
				assert.Equal(t, resourceazure.AzureNetworkSecurityGroupResourceType, got[1].ResourceType())
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockNetworkRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.NetworkRepository = fakeRepo
			providerVersion := "2.71.0"
			realProvider, err := terraform2.InitTestAzureProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
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
				repo = repository.NewNetworkRepository(cred, clientOptions, realProvider.GetConfig(), cache.New(0))
			}

			remoteLibrary.AddEnumerator(azurerm.NewAzurermNetworkSecurityGroupEnumerator(repo, factory))

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

func TestAzurermLoadBalancers(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockNetworkRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no load balancer",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return([]*armnetwork.LoadBalancer{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "error listing load balancers",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return(nil, dummyError)
			},
			wantErr: error2.NewResourceListingError(dummyError, resourceazure.AzureLoadBalancerResourceType),
		},
		{
			test: "multiple load balancers",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return([]*armnetwork.LoadBalancer{
					{
						Resource: armnetwork.Resource{
							ID:   to.StringPtr("lb-1"), // Here we don't care to have a valid ID, it is for testing purpose only
							Name: to.StringPtr("lb-1"),
						},
					},
					{
						Resource: armnetwork.Resource{
							ID:   to.StringPtr("lb-2"),
							Name: to.StringPtr("lb-2"),
						},
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, got[0].ResourceId(), "lb-1")
				assert.Equal(t, got[0].ResourceType(), resourceazure.AzureLoadBalancerResourceType)

				assert.Equal(t, got[1].ResourceId(), "lb-2")
				assert.Equal(t, got[1].ResourceType(), resourceazure.AzureLoadBalancerResourceType)
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockNetworkRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.NetworkRepository = fakeRepo

			remoteLibrary.AddEnumerator(azurerm.NewAzurermLoadBalancerEnumerator(repo, factory))

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

func TestAzurermLoadBalancerRules(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockNetworkRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			test:    "no load balancer rule",
			dirName: "azurerm_lb_rule_empty",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				loadbalancer := &armnetwork.LoadBalancer{
					Resource: armnetwork.Resource{
						ID:   to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/raphael-dev/providers/Microsoft.Network/loadBalancers/TestLoadBalancer/frontendIPConfigurations/PublicIPAddress"),
						Name: to.StringPtr("testlb"),
					},
				}

				repository.On("ListAllLoadBalancers").Return([]*armnetwork.LoadBalancer{loadbalancer}, nil)

				repository.On("ListLoadBalancerRules", loadbalancer).Return([]*armnetwork.LoadBalancingRule{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test:    "error listing load balancer rules",
			dirName: "azurerm_lb_rule_empty",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return(nil, dummyError)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: error2.NewResourceListingErrorWithType(dummyError, resourceazure.AzureLoadBalancerRuleResourceType, resourceazure.AzureLoadBalancerResourceType),
		},
		{
			test:    "multiple load balancer rules",
			dirName: "azurerm_lb_rule_multiple",
			mocks: func(repository *repository.MockNetworkRepository, alerter *mocks.AlerterInterface) {
				loadbalancer := &armnetwork.LoadBalancer{
					Resource: armnetwork.Resource{
						ID:   to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/raphael-dev/providers/Microsoft.Network/loadBalancers/TestLoadBalancer/frontendIPConfigurations/PublicIPAddress"),
						Name: to.StringPtr("TestLoadBalancer"),
					},
				}

				repository.On("ListAllLoadBalancers").Return([]*armnetwork.LoadBalancer{loadbalancer}, nil)

				repository.On("ListLoadBalancerRules", loadbalancer).Return([]*armnetwork.LoadBalancingRule{
					{
						SubResource: armnetwork.SubResource{
							ID: to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/raphael-dev/providers/Microsoft.Network/loadBalancers/TestLoadBalancer/loadBalancingRules/LBRule"),
						},
						Name: to.StringPtr("LBRule"),
					},
					{
						SubResource: armnetwork.SubResource{
							ID: to.StringPtr("/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/raphael-dev/providers/Microsoft.Network/loadBalancers/TestLoadBalancer/loadBalancingRules/LBRule2"),
						},
						Name: to.StringPtr("LBRule2"),
					},
				}, nil).Once()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/raphael-dev/providers/Microsoft.Network/loadBalancers/TestLoadBalancer/loadBalancingRules/LBRule", got[0].ResourceId())
				assert.Equal(t, resourceazure.AzureLoadBalancerRuleResourceType, got[0].ResourceType())

				assert.Equal(t, "/subscriptions/8cb43347-a79f-4bb2-a8b4-c838b41fa5a5/resourceGroups/raphael-dev/providers/Microsoft.Network/loadBalancers/TestLoadBalancer/loadBalancingRules/LBRule2", got[1].ResourceId())
				assert.Equal(t, resourceazure.AzureLoadBalancerRuleResourceType, got[1].ResourceType())
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockNetworkRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.NetworkRepository = fakeRepo
			providerVersion := "2.71.0"
			realProvider, err := terraform2.InitTestAzureProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
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
				repo = repository.NewNetworkRepository(cred, clientOptions, realProvider.GetConfig(), cache.New(0))
			}

			remoteLibrary.AddEnumerator(azurerm.NewAzurermLoadBalancerRuleEnumerator(repo, factory))

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

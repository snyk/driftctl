package azurerm

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/azurerm"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

func Init(
	version string,
	alerter *alerter.Alerter,
	providerLibrary *terraform.ProviderLibrary,
	remoteLibrary *common.RemoteLibrary,
	progress output.Progress,
	resourceSchemaRepository *resource.SchemaRepository,
	factory resource.ResourceFactory,
	configDir string) error {

	provider, err := NewAzureTerraformProvider(version, progress, configDir)
	if err != nil {
		return err
	}
	err = provider.Init()
	if err != nil {
		return err
	}

	providerConfig := provider.GetConfig()
	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{})
	if err != nil {
		return err
	}
	con := arm.NewDefaultConnection(cred, nil)

	c := cache.New(100)

	storageAccountRepo := repository.NewStorageRepository(con, providerConfig, c)
	networkRepo := repository.NewNetworkRepository(con, providerConfig, c)
	resourcesRepo := repository.NewResourcesRepository(con, providerConfig, c)
	containerRegistryRepo := repository.NewContainerRegistryRepository(con, providerConfig, c)
	postgresqlRepo := repository.NewPostgresqlRepository(con, providerConfig, c)
	privateDNSRepo := repository.NewPrivateDNSRepository(con, providerConfig, c)
	computeRepo := repository.NewComputeRepository(con, providerConfig, c)

	providerLibrary.AddProvider(terraform.AZURE, provider)
	deserializer := resource.NewDeserializer(factory)

	remoteLibrary.AddEnumerator(NewAzurermStorageAccountEnumerator(storageAccountRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermStorageContainerEnumerator(storageAccountRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermVirtualNetworkEnumerator(networkRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermRouteTableEnumerator(networkRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermRouteEnumerator(networkRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermResourceGroupEnumerator(resourcesRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermSubnetEnumerator(networkRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermContainerRegistryEnumerator(containerRegistryRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermFirewallsEnumerator(networkRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermPostgresqlServerEnumerator(postgresqlRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermPublicIPEnumerator(networkRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermPostgresqlDatabaseEnumerator(postgresqlRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermNetworkSecurityGroupEnumerator(networkRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzureNetworkSecurityGroupResourceType, common.NewGenericDetailsFetcher(azurerm.AzureNetworkSecurityGroupResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewAzurermLoadBalancerEnumerator(networkRepo, factory))

	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSZoneEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzurePrivateDNSZoneResourceType, common.NewGenericDetailsFetcher(azurerm.AzurePrivateDNSZoneResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSARecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzurePrivateDNSARecordResourceType, common.NewGenericDetailsFetcher(azurerm.AzurePrivateDNSARecordResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSAAAARecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzurePrivateDNSAAAARecordResourceType, common.NewGenericDetailsFetcher(azurerm.AzurePrivateDNSAAAARecordResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewAzurermImageEnumerator(computeRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermSSHPublicKeyEnumerator(computeRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzureSSHPublicKeyResourceType, common.NewGenericDetailsFetcher(azurerm.AzureSSHPublicKeyResourceType, provider, deserializer))

	err = resourceSchemaRepository.Init(terraform.AZURE, provider.Version(), provider.Schema())
	if err != nil {
		return err
	}
	azurerm.InitResourcesMetadata(resourceSchemaRepository)

	return nil
}

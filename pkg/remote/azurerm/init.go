package azurerm

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/snyk/driftctl/pkg/alerter"
	"github.com/snyk/driftctl/pkg/output"
	"github.com/snyk/driftctl/pkg/remote/azurerm/repository"
	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/snyk/driftctl/pkg/remote/common"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/azurerm"
	"github.com/snyk/driftctl/pkg/terraform"
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
	err = provider.CheckCredentialsExist()
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
	clientOptions := &arm.ClientOptions{}

	c := cache.New(100)

	storageAccountRepo := repository.NewStorageRepository(cred, clientOptions, providerConfig, c)
	networkRepo := repository.NewNetworkRepository(cred, clientOptions, providerConfig, c)
	resourcesRepo := repository.NewResourcesRepository(cred, clientOptions, providerConfig, c)
	containerRegistryRepo := repository.NewContainerRegistryRepository(cred, clientOptions, providerConfig, c)
	postgresqlRepo := repository.NewPostgresqlRepository(cred, clientOptions, providerConfig, c)
	privateDNSRepo := repository.NewPrivateDNSRepository(cred, clientOptions, providerConfig, c)
	computeRepo := repository.NewComputeRepository(cred, clientOptions, providerConfig, c)

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
	remoteLibrary.AddEnumerator(NewAzurermLoadBalancerRuleEnumerator(networkRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzureLoadBalancerRuleResourceType, common.NewGenericDetailsFetcher(azurerm.AzureLoadBalancerRuleResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSZoneEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzurePrivateDNSZoneResourceType, common.NewGenericDetailsFetcher(azurerm.AzurePrivateDNSZoneResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSARecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzurePrivateDNSARecordResourceType, common.NewGenericDetailsFetcher(azurerm.AzurePrivateDNSARecordResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSAAAARecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzurePrivateDNSAAAARecordResourceType, common.NewGenericDetailsFetcher(azurerm.AzurePrivateDNSAAAARecordResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSMXRecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzurePrivateDNSMXRecordResourceType, common.NewGenericDetailsFetcher(azurerm.AzurePrivateDNSMXRecordResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSCNameRecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzurePrivateDNSCNameRecordResourceType, common.NewGenericDetailsFetcher(azurerm.AzurePrivateDNSCNameRecordResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSPTRRecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzurePrivateDNSPTRRecordResourceType, common.NewGenericDetailsFetcher(azurerm.AzurePrivateDNSPTRRecordResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSSRVRecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzurePrivateDNSSRVRecordResourceType, common.NewGenericDetailsFetcher(azurerm.AzurePrivateDNSSRVRecordResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSTXTRecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddDetailsFetcher(azurerm.AzurePrivateDNSTXTRecordResourceType, common.NewGenericDetailsFetcher(azurerm.AzurePrivateDNSTXTRecordResourceType, provider, deserializer))

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

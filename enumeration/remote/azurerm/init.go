package azurerm

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/azurerm/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/terraform"
)

func Init(version string, alerter alerter.AlerterInterface, providerLibrary *terraform.ProviderLibrary, remoteLibrary *common.RemoteLibrary, progress enumeration.ProgressCounter, factory resource.ResourceFactory, configDir string) error {

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
	remoteLibrary.AddEnumerator(NewAzurermLoadBalancerEnumerator(networkRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermLoadBalancerRuleEnumerator(networkRepo, factory))

	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSZoneEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSARecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSAAAARecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSMXRecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSCNameRecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSPTRRecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSSRVRecordEnumerator(privateDNSRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermPrivateDNSTXTRecordEnumerator(privateDNSRepo, factory))

	remoteLibrary.AddEnumerator(NewAzurermImageEnumerator(computeRepo, factory))
	remoteLibrary.AddEnumerator(NewAzurermSSHPublicKeyEnumerator(computeRepo, factory))

	return nil
}

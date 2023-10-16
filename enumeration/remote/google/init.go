package google

import (
	"context"

	"github.com/snyk/driftctl/enumeration"

	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/remote/google/repository"
	"github.com/snyk/driftctl/enumeration/terraform"

	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/storage"
	"github.com/snyk/driftctl/enumeration/resource"
	"google.golang.org/api/cloudresourcemanager/v1"
)

func Init(version string, alerter alerter.AlerterInterface, providerLibrary *terraform.ProviderLibrary, remoteLibrary *common.RemoteLibrary, progress enumeration.ProgressCounter, factory resource.ResourceFactory, configDir string) error {

	provider, err := NewGCPTerraformProvider(version, progress, configDir)
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

	repositoryCache := cache.New(100)

	ctx := context.Background()
	assetClient, err := asset.NewClient(ctx)
	if err != nil {
		return err
	}

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	crmService, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return err
	}

	assetRepository := repository.NewAssetRepository(assetClient, provider.GetConfig(), repositoryCache)
	storageRepository := repository.NewStorageRepository(storageClient, repositoryCache)
	iamRepository := repository.NewCloudResourceManagerRepository(crmService, provider.GetConfig(), repositoryCache)

	providerLibrary.AddProvider(terraform.GOOGLE, provider)

	remoteLibrary.AddEnumerator(NewGoogleStorageBucketEnumerator(assetRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleComputeFirewallEnumerator(assetRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleComputeRouterEnumerator(assetRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleComputeInstanceEnumerator(assetRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleProjectIamMemberEnumerator(iamRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleStorageBucketIamMemberEnumerator(assetRepository, storageRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleComputeNetworkEnumerator(assetRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleComputeSubnetworkEnumerator(assetRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleDNSManagedZoneEnumerator(assetRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleComputeInstanceGroupEnumerator(assetRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleBigqueryDatasetEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleBigqueryTableEnumerator(assetRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleComputeAddressEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleComputeGlobalAddressEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleCloudFunctionsFunctionEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleComputeDiskEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleComputeImageEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleBigTableInstanceEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleBigtableTableEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleSQLDatabaseInstanceEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleComputeHealthCheckEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleCloudRunServiceEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleComputeNodeGroupEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleComputeForwardingRuleEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleComputeInstanceGroupManagerEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleComputeGlobalForwardingRuleEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleComputeSslCertificateEnumerator(assetRepository, factory))

	return nil
}

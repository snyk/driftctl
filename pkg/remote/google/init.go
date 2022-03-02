package google

import (
	"context"

	asset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/storage"
	"github.com/snyk/driftctl/pkg/alerter"
	"github.com/snyk/driftctl/pkg/output"
	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/snyk/driftctl/pkg/remote/common"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
	"github.com/snyk/driftctl/pkg/terraform"
	"google.golang.org/api/cloudresourcemanager/v1"
)

func Init(version string, gcpScope []string, alerter *alerter.Alerter,
	providerLibrary *terraform.ProviderLibrary,
	remoteLibrary *common.RemoteLibrary,
	progress output.Progress,
	resourceSchemaRepository *resource.SchemaRepository,
	factory resource.ResourceFactory,
	configDir string) error {

	provider, err := NewGCPTerraformProvider(version, progress, configDir)
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

	assetRepository := repository.NewAssetRepository(assetClient, provider.SetConfig(gcpScope), repositoryCache)
	storageRepository := repository.NewStorageRepository(storageClient, repositoryCache)
	iamRepository := repository.NewCloudResourceManagerRepository(crmService, provider.SetConfig(gcpScope), repositoryCache)

	providerLibrary.AddProvider(terraform.GOOGLE, provider)
	deserializer := resource.NewDeserializer(factory)

	remoteLibrary.AddEnumerator(NewGoogleStorageBucketEnumerator(assetRepository, factory))
	remoteLibrary.AddDetailsFetcher(google.GoogleStorageBucketResourceType, common.NewGenericDetailsFetcher(google.GoogleStorageBucketResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewGoogleComputeFirewallEnumerator(assetRepository, factory))
	remoteLibrary.AddDetailsFetcher(google.GoogleComputeFirewallResourceType, common.NewGenericDetailsFetcher(google.GoogleComputeFirewallResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewGoogleComputeRouterEnumerator(assetRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleComputeInstanceEnumerator(assetRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleProjectIamMemberEnumerator(iamRepository, factory))
	remoteLibrary.AddDetailsFetcher(google.GoogleProjectIamMemberResourceType, common.NewGenericDetailsFetcher(google.GoogleProjectIamMemberResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewGoogleStorageBucketIamMemberEnumerator(assetRepository, storageRepository, factory))
	remoteLibrary.AddDetailsFetcher(google.GoogleStorageBucketIamMemberResourceType, common.NewGenericDetailsFetcher(google.GoogleStorageBucketIamMemberResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewGoogleComputeNetworkEnumerator(assetRepository, factory))
	remoteLibrary.AddDetailsFetcher(google.GoogleComputeNetworkResourceType, common.NewGenericDetailsFetcher(google.GoogleComputeNetworkResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewGoogleComputeSubnetworkEnumerator(assetRepository, factory))
	remoteLibrary.AddDetailsFetcher(google.GoogleComputeSubnetworkResourceType, common.NewGenericDetailsFetcher(google.GoogleComputeSubnetworkResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewGoogleDNSManagedZoneEnumerator(assetRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleComputeInstanceGroupEnumerator(assetRepository, factory))
	remoteLibrary.AddDetailsFetcher(google.GoogleComputeInstanceGroupResourceType, common.NewGenericDetailsFetcher(google.GoogleComputeInstanceGroupResourceType, provider, deserializer))

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

	err = resourceSchemaRepository.Init(terraform.GOOGLE, provider.Version(), provider.Schema())
	if err != nil {
		return err
	}
	google.InitResourcesMetadata(resourceSchemaRepository)

	return nil
}

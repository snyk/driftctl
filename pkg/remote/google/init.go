package google

import (
	"context"

	asset "cloud.google.com/go/asset/apiv1"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	"github.com/cloudskiff/driftctl/pkg/remote/google/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

func Init(version string, alerter *alerter.Alerter,
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
	assetRepository := repository.NewAssetRepository(assetClient, provider.GetConfig(), repositoryCache)

	storageRepository := repository.NewStorageRepository(repositoryCache)

	providerLibrary.AddProvider(terraform.GOOGLE, provider)
	deserializer := resource.NewDeserializer(factory)

	remoteLibrary.AddEnumerator(NewGoogleStorageBucketEnumerator(assetRepository, factory))
	remoteLibrary.AddDetailsFetcher(google.GoogleStorageBucketResourceType, common.NewGenericDetailsFetcher(google.GoogleStorageBucketResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewGoogleComputeFirewallEnumerator(assetRepository, factory))
	remoteLibrary.AddDetailsFetcher(google.GoogleComputeFirewallResourceType, common.NewGenericDetailsFetcher(google.GoogleComputeFirewallResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewGoogleComputeRouterEnumerator(assetRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleComputeInstanceEnumerator(assetRepository, factory))
	remoteLibrary.AddEnumerator(NewGoogleStorageBucketIamBindingEnumerator(assetRepository, storageRepository, factory))

	remoteLibrary.AddEnumerator(NewGoogleComputeNetworkEnumerator(assetRepository, factory))
	remoteLibrary.AddDetailsFetcher(google.GoogleComputeNetworkResourceType, common.NewGenericDetailsFetcher(google.GoogleComputeNetworkResourceType, provider, deserializer))

	err = resourceSchemaRepository.Init(terraform.GOOGLE, provider.Version(), provider.Schema())
	if err != nil {
		return err
	}
	google.InitResourcesMetadata(resourceSchemaRepository)

	return nil
}

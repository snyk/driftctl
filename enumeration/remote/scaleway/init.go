package scaleway

import (
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/remote/scaleway/client"
	"github.com/snyk/driftctl/enumeration/remote/scaleway/repository"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/terraform"
	"github.com/snyk/driftctl/pkg/resource/scaleway"
)

func Init(version string, alerter alerter.AlerterInterface, providerLibrary *terraform.ProviderLibrary, remoteLibrary *common.RemoteLibrary, progress enumeration.ProgressCounter, factory resource.ResourceFactory, configDir string) error {

	if version == "" {
		version = "2.14.1"
	}

	provider, err := NewScalewayTerraformProvider(version, progress, configDir)
	if err != nil {
		return err
	}
	err = provider.Init()
	if err != nil {
		return err
	}

	providerLibrary.AddProvider(terraform.SCALEWAY, provider)

	scwClient, err := client.Create()
	if err != nil {
		return err
	}

	repositoryCache := cache.New(100)

	funcRepository := repository.NewFunctionRepository(scwClient, repositoryCache)

	deserializer := resource.NewDeserializer(factory)

	remoteLibrary.AddEnumerator(NewFunctionNamespaceEnumerator(funcRepository, factory))
	remoteLibrary.AddDetailsFetcher(scaleway.ScalewayFunctionNamespaceResourceType, common.NewGenericDetailsFetcher(scaleway.ScalewayFunctionNamespaceResourceType, provider, deserializer))

	return nil
}

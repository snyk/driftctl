package github

import (
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/github"
	"github.com/snyk/driftctl/enumeration/terraform"
)

/**
 * Initialize remote (configure credentials, launch tf providers and start gRPC clients)
 * Required to use Scanner
 */

func Init(version string, alerter *alerter.Alerter,
	providerLibrary *terraform.ProviderLibrary,
	remoteLibrary *common.RemoteLibrary,
	progress enumeration.ProgressCounter,
	resourceSchemaRepository *resource.SchemaRepository,
	factory resource.ResourceFactory,
	configDir string) error {

	provider, err := NewGithubTerraformProvider(version, progress, configDir)
	if err != nil {
		return err
	}
	err = provider.Init()
	if err != nil {
		return err
	}

	repositoryCache := cache.New(100)

	repository := NewGithubRepository(provider.GetConfig(), repositoryCache)
	deserializer := resource.NewDeserializer(factory)
	providerLibrary.AddProvider(terraform.GITHUB, provider)

	remoteLibrary.AddEnumerator(NewGithubTeamEnumerator(repository, factory))
	remoteLibrary.AddDetailsFetcher(github.GithubTeamResourceType, common.NewGenericDetailsFetcher(github.GithubTeamResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewGithubRepositoryEnumerator(repository, factory))
	remoteLibrary.AddDetailsFetcher(github.GithubRepositoryResourceType, common.NewGenericDetailsFetcher(github.GithubRepositoryResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewGithubMembershipEnumerator(repository, factory))
	remoteLibrary.AddDetailsFetcher(github.GithubMembershipResourceType, common.NewGenericDetailsFetcher(github.GithubMembershipResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewGithubTeamMembershipEnumerator(repository, factory))
	remoteLibrary.AddDetailsFetcher(github.GithubTeamMembershipResourceType, common.NewGenericDetailsFetcher(github.GithubTeamMembershipResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewGithubBranchProtectionEnumerator(repository, factory))
	remoteLibrary.AddDetailsFetcher(github.GithubBranchProtectionResourceType, common.NewGenericDetailsFetcher(github.GithubBranchProtectionResourceType, provider, deserializer))

	err = resourceSchemaRepository.Init(terraform.GITHUB, provider.Version(), provider.Schema())
	if err != nil {
		return err
	}
	github.InitResourcesMetadata(resourceSchemaRepository)

	return nil
}

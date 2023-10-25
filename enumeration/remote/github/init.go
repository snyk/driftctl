package github

import (
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/terraform"
)

/**
 * Initialize remote (configure credentials, launch tf providers and start gRPC clients)
 * Required to use Scanner
 */

func Init(version string, alerter alerter.AlerterInterface, providerLibrary *terraform.ProviderLibrary, remoteLibrary *common.RemoteLibrary, progress enumeration.ProgressCounter, factory resource.ResourceFactory, configDir string) error {

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
	providerLibrary.AddProvider(terraform.GITHUB, provider)

	remoteLibrary.AddEnumerator(NewGithubTeamEnumerator(repository, factory))

	remoteLibrary.AddEnumerator(NewGithubRepositoryEnumerator(repository, factory))

	remoteLibrary.AddEnumerator(NewGithubMembershipEnumerator(repository, factory))

	remoteLibrary.AddEnumerator(NewGithubTeamMembershipEnumerator(repository, factory))

	remoteLibrary.AddEnumerator(NewGithubBranchProtectionEnumerator(repository, factory))

	return nil
}

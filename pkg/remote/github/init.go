package github

import (
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/github"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

const RemoteGithubTerraform = "github+tf"

/**
 * Initialize remote (configure credentials, launch tf providers and start gRPC clients)
 * Required to use Scanner
 */

func Init(version string, alerter *alerter.Alerter,
	providerLibrary *terraform.ProviderLibrary,
	supplierLibrary *resource.SupplierLibrary,
	progress output.Progress,
	resourceSchemaRepository *resource.SchemaRepository,
	factory resource.ResourceFactory,
	ignore *filter.DriftIgnore) error {

	provider, err := NewGithubTerraformProvider(version, progress, ignore)
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

	var githubSuppliers []resource.SimpleTypeSupplier
	githubSuppliers = append(githubSuppliers, NewGithubRepositorySupplier(provider, repository, deserializer))
	githubSuppliers = append(githubSuppliers, NewGithubTeamSupplier(provider, repository, deserializer))
	githubSuppliers = append(githubSuppliers, NewGithubMembershipSupplier(provider, repository, deserializer))
	githubSuppliers = append(githubSuppliers, NewGithubTeamMembershipSupplier(provider, repository, deserializer))
	githubSuppliers = append(githubSuppliers, NewGithubBranchProtectionSupplier(provider, repository, deserializer))

	for _, supplier := range githubSuppliers {
		if ignore.IsTypeIgnored(supplier.SuppliedType().String()) {
			continue
		}
		supplierLibrary.AddSupplier(supplier)
	}

	err = resourceSchemaRepository.Init(version, provider.Schema())
	if err != nil {
		return err
	}
	github.InitResourcesMetadata(resourceSchemaRepository)

	return nil
}

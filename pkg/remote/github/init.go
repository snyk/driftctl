package github

import (
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/output"
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
	factory resource.ResourceFactory) error {

	provider, err := NewGithubTerraformProvider(version, progress)
	if err != nil {
		return err
	}
	err = provider.Init()
	if err != nil {
		return err
	}

	repository := NewGithubRepository(provider.GetConfig())
	deserializer := resource.NewDeserializer(factory)
	providerLibrary.AddProvider(terraform.GITHUB, provider)

	supplierLibrary.AddSupplier(NewGithubRepositorySupplier(provider, repository, deserializer))
	supplierLibrary.AddSupplier(NewGithubTeamSupplier(provider, repository, deserializer))
	supplierLibrary.AddSupplier(NewGithubMembershipSupplier(provider, repository, deserializer))
	supplierLibrary.AddSupplier(NewGithubTeamMembershipSupplier(provider, repository, deserializer))
	supplierLibrary.AddSupplier(NewGithubBranchProtectionSupplier(provider, repository, deserializer))

	resourceSchemaRepository.Init(provider.Schema())
	github.InitResourcesMetadata(resourceSchemaRepository)

	return nil
}

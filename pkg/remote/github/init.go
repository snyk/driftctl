package github

import (
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

const RemoteGithubTerraform = "github+tf"

/**
 * Initialize remote (configure credentials, launch tf providers and start gRPC clients)
 * Required to use Scanner
 */
func Init(alerter *alerter.Alerter, providerLibrary *terraform.ProviderLibrary, supplierLibrary *resource.SupplierLibrary) error {
	provider, err := NewGithubTerraformProvider()
	if err != nil {
		return err
	}
	err = provider.Init()
	if err != nil {
		return err
	}

	repository := NewGithubRepository(provider.GetConfig())

	providerLibrary.AddProvider(terraform.GITHUB, provider)

	supplierLibrary.AddSupplier(NewGithubRepositorySupplier(provider, repository))
	supplierLibrary.AddSupplier(NewGithubTeamSupplier(provider, repository))
	supplierLibrary.AddSupplier(NewGithubMembershipSupplier(provider, repository))

	return nil
}

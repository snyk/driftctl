package azurerm

import (
	"os"

	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/common"
	"github.com/cloudskiff/driftctl/pkg/remote/terraform"
	tf "github.com/cloudskiff/driftctl/pkg/terraform"
)

type AzureTerraformProvider struct {
	*terraform.TerraformProvider
	name    string
	version string
}

func NewAzureTerraformProvider(version string, progress output.Progress, configDir string) (*AzureTerraformProvider, error) {
	// Just pass your version and name
	p := &AzureTerraformProvider{
		version: version,
		name:    tf.AZURE,
	}
	// Use TerraformProviderInstaller to retrieve the provider if needed
	installer, err := tf.NewProviderInstaller(tf.ProviderConfig{
		Key:       p.name,
		Version:   version,
		ConfigDir: configDir,
	})
	if err != nil {
		return nil, err
	}

	tfProvider, err := terraform.NewTerraformProvider(installer, terraform.TerraformProviderConfig{
		Name: p.name,
	}, progress)
	if err != nil {
		return nil, err
	}
	p.TerraformProvider = tfProvider
	return p, err
}

func (p *AzureTerraformProvider) GetConfig() common.AzureProviderConfig {
	return common.AzureProviderConfig{
		SubscriptionID: os.Getenv("AZURE_SUBSCRIPTION_ID"),
	}
}

func (p *AzureTerraformProvider) Name() string {
	return p.name
}

func (p *AzureTerraformProvider) Version() string {
	return p.version
}

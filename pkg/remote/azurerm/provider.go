package azurerm

import (
	"os"

	"github.com/snyk/driftctl/pkg/output"
	"github.com/snyk/driftctl/pkg/remote/azurerm/common"
	"github.com/snyk/driftctl/pkg/remote/terraform"
	tf "github.com/snyk/driftctl/pkg/terraform"
)

type AzureTerraformProvider struct {
	*terraform.TerraformProvider
	name    string
	version string
}

func NewAzureTerraformProvider(version string, progress output.Progress, configDir string) (*AzureTerraformProvider, error) {
	if version == "" {
		version = "2.71.0"
	}
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
		GetProviderConfig: func(_ string) interface{} {
			c := p.GetConfig()
			return map[string]interface{}{
				"subscription_id":            c.SubscriptionID,
				"tenant_id":                  c.TenantID,
				"client_id":                  c.ClientID,
				"client_secret":              c.ClientSecret,
				"skip_provider_registration": true,
			}
		},
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
		TenantID:       os.Getenv("AZURE_TENANT_ID"),
		ClientID:       os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret:   os.Getenv("AZURE_CLIENT_SECRET"),
	}
}

func (p *AzureTerraformProvider) Name() string {
	return p.name
}

func (p *AzureTerraformProvider) Version() string {
	return p.version
}

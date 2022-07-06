package azurerm

import (
	"context"
	"errors"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/azurerm/common"
	"github.com/snyk/driftctl/enumeration/remote/terraform"
	tf "github.com/snyk/driftctl/enumeration/terraform"
)

type AzureTerraformProvider struct {
	*terraform.TerraformProvider
	name    string
	version string
}

func NewAzureTerraformProvider(version string, progress enumeration.ProgressCounter, configDir string) (*AzureTerraformProvider, error) {
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

func (p *AzureTerraformProvider) CheckCredentialsExist() error {
	cred, err := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{})
	if err != nil {
		return err
	}

	_, err = cred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{"https://management.azure.com//.default"}})
	if err != nil {
		return errors.New("Could not find any authentication method for Azure.\n" +
			"For more information, please check the official Azure documentation: https://docs.microsoft.com/en-us/azure/developer/go/azure-sdk-authorization#use-environment-based-authentication")
	}

	if p.GetConfig().SubscriptionID == "" {
		return errors.New("Please provide an Azure subscription ID by setting the `AZURE_SUBSCRIPTION_ID` environment variable.")
	}

	return nil
}

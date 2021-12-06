package terraform

import (
	"os"

	"github.com/snyk/driftctl/pkg/output"
	"github.com/snyk/driftctl/pkg/remote/aws"
	"github.com/snyk/driftctl/pkg/remote/azurerm"
	"github.com/snyk/driftctl/pkg/remote/github"
	"github.com/snyk/driftctl/pkg/remote/google"
	"github.com/snyk/driftctl/pkg/terraform"
)

func InitTestAwsProvider(providerLibrary *terraform.ProviderLibrary, version string) (*aws.AWSTerraformProvider, error) {
	progress := &output.MockProgress{}
	progress.On("Inc").Maybe().Return()
	provider, err := aws.NewAWSTerraformProvider(version, progress, os.TempDir())
	if err != nil {
		return nil, err
	}
	providerLibrary.AddProvider(terraform.AWS, provider)
	return provider, nil
}

func InitTestGithubProvider(providerLibrary *terraform.ProviderLibrary, version string) (*github.GithubTerraformProvider, error) {
	progress := &output.MockProgress{}
	progress.On("Inc").Maybe().Return()
	provider, err := github.NewGithubTerraformProvider(version, progress, os.TempDir())
	if err != nil {
		return nil, err
	}
	providerLibrary.AddProvider(terraform.GITHUB, provider)

	return provider, nil
}

func InitTestGoogleProvider(providerLibrary *terraform.ProviderLibrary, version string) (*google.GCPTerraformProvider, error) {
	progress := &output.MockProgress{}
	progress.On("Inc").Maybe().Return()
	provider, err := google.NewGCPTerraformProvider(version, progress, os.TempDir())
	if err != nil {
		return nil, err
	}
	providerLibrary.AddProvider(terraform.GOOGLE, provider)

	return provider, nil
}

func InitTestAzureProvider(providerLibrary *terraform.ProviderLibrary, version string) (*azurerm.AzureTerraformProvider, error) {
	progress := &output.MockProgress{}
	progress.On("Inc").Maybe().Return()
	provider, err := azurerm.NewAzureTerraformProvider(version, progress, os.TempDir())
	if err != nil {
		return nil, err
	}
	providerLibrary.AddProvider(terraform.AZURE, provider)

	return provider, nil
}

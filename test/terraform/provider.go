package terraform

import (
	"os"

	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/github"
	"github.com/cloudskiff/driftctl/pkg/remote/google"
	"github.com/cloudskiff/driftctl/pkg/terraform"
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
	provider, err := github.NewGithubTerraformProvider(version, progress, "")
	if err != nil {
		return nil, err
	}
	providerLibrary.AddProvider(terraform.GITHUB, provider)

	return provider, nil
}

func InitTestGoogleProvider(providerLibrary *terraform.ProviderLibrary, version string) (*google.GCPTerraformProvider, error) {
	progress := &output.MockProgress{}
	progress.On("Inc").Maybe().Return()
	provider, err := google.NewGCPTerraformProvider(version, progress, "")
	if err != nil {
		return nil, err
	}
	providerLibrary.AddProvider(terraform.GOOGLE, provider)

	return provider, nil
}

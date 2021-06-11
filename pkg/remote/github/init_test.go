package github

import (
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

func InitTestGithubProvider(providerLibrary *terraform.ProviderLibrary) (*GithubTerraformProvider, error) {
	provider, err := NewGithubTerraformProvider("", &output.MockProgress{}, filter.NewDriftIgnore())
	if err != nil {
		return nil, err
	}
	err = provider.Init()
	if err != nil {
		return nil, err
	}
	providerLibrary.AddProvider(terraform.GITHUB, provider)

	return provider, nil
}

package github

import (
	"os"

	"github.com/cloudskiff/driftctl/pkg/output"

	"github.com/cloudskiff/driftctl/pkg/remote/terraform"
	tf "github.com/cloudskiff/driftctl/pkg/terraform"
)

type GithubTerraformProvider struct {
	*terraform.TerraformProvider
}

type githubConfig struct {
	Token        string
	Owner        string `cty:"owner"`
	Organization string
}

func NewGithubTerraformProvider(version string, progress output.Progress) (*GithubTerraformProvider, error) {
	p := &GithubTerraformProvider{}
	providerKey := "github"
	if version == "" {
		version = "4.4.0"
	}
	installer, err := tf.NewProviderInstaller(tf.ProviderConfig{
		Key:     providerKey,
		Version: version,
	})
	if err != nil {
		return nil, err
	}
	tfProvider, err := terraform.NewTerraformProvider(installer, terraform.TerraformProviderConfig{
		Name:         providerKey,
		DefaultAlias: p.GetConfig().getDefaultOwner(),
		GetProviderConfig: func(owner string) interface{} {
			return githubConfig{
				Owner: p.GetConfig().getDefaultOwner(),
			}
		},
	}, progress)
	if err != nil {
		return nil, err
	}
	p.TerraformProvider = tfProvider
	return p, err
}

func (c githubConfig) getDefaultOwner() string {
	if c.Organization != "" {
		return c.Organization
	}
	return c.Owner
}

func (p GithubTerraformProvider) GetConfig() githubConfig {
	return githubConfig{
		Token:        os.Getenv("GITHUB_TOKEN"),
		Owner:        os.Getenv("GITHUB_OWNER"),
		Organization: os.Getenv("GITHUB_ORGANIZATION"),
	}
}

package scaleway

import (
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/terraform"
	tf "github.com/snyk/driftctl/enumeration/terraform"
)

type ScalewayTerraformProvider struct {
	*terraform.TerraformProvider
	name    string
	version string
}

func NewScalewayTerraformProvider(version string, progress enumeration.ProgressCounter, configDir string) (*ScalewayTerraformProvider, error) {

	provider := &ScalewayTerraformProvider{
		version: version,
		name:    "scaleway",
	}

	installer, err := tf.NewProviderInstaller(tf.ProviderConfig{
		Key:       provider.name,
		Version:   version,
		ConfigDir: configDir,
	})
	if err != nil {
		return nil, err
	}

	tfProvider, err := terraform.NewTerraformProvider(installer, terraform.TerraformProviderConfig{
		Name: provider.name,
	}, progress)
	if err != nil {
		return nil, err
	}
	provider.TerraformProvider = tfProvider
	return provider, err
}

func (p *ScalewayTerraformProvider) Name() string {
	return p.name
}

func (p *ScalewayTerraformProvider) Version() string {
	return p.version
}

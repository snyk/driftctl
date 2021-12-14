package google

import (
	"os"

	"github.com/snyk/driftctl/pkg/output"
	"github.com/snyk/driftctl/pkg/remote/google/config"
	"github.com/snyk/driftctl/pkg/remote/terraform"
	tf "github.com/snyk/driftctl/pkg/terraform"
)

type GCPTerraformProvider struct {
	*terraform.TerraformProvider
	name    string
	version string
}

func NewGCPTerraformProvider(version string, progress output.Progress, configDir string) (*GCPTerraformProvider, error) {
	if version == "" {
		version = "3.78.0"
	}
	p := &GCPTerraformProvider{
		version: version,
		name:    tf.GOOGLE,
	}
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
		GetProviderConfig: func(alias string) interface{} {
			return p.GetConfig()
		},
	}, progress)

	if err != nil {
		return nil, err
	}

	p.TerraformProvider = tfProvider

	return p, err
}

func (p *GCPTerraformProvider) Name() string {
	return p.name
}

func (p *GCPTerraformProvider) Version() string {
	return p.version
}

func (p *GCPTerraformProvider) GetConfig() config.GCPTerraformConfig {
	return config.GCPTerraformConfig{
		Organization: os.Getenv("CLOUDSDK_ORGANIZATION"),
		Project:      os.Getenv("CLOUDSDK_CORE_PROJECT"),
		Region:       os.Getenv("CLOUDSDK_COMPUTE_REGION"),
		Zone:         os.Getenv("CLOUDSDK_COMPUTE_ZONE"),
	}
}

func (p *GCPTerraformProvider) SetConfig(organization string, project string, region string, zone string) config.GCPTerraformConfig {
	return config.GCPTerraformConfig{
		Organization: organization,
		Project:      project,
		Region:       region,
		Zone:         zone,
	}
}

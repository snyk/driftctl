package google

import (
	"context"
	"errors"
	"os"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/google/config"
	"github.com/snyk/driftctl/enumeration/remote/terraform"
	terraform2 "github.com/snyk/driftctl/enumeration/terraform"

	asset "cloud.google.com/go/asset/apiv1"
)

type GCPTerraformProvider struct {
	*terraform.TerraformProvider
	name    string
	version string
}

func NewGCPTerraformProvider(version string, progress enumeration.ProgressCounter, configDir string) (*GCPTerraformProvider, error) {
	if version == "" {
		version = "3.78.0"
	}
	p := &GCPTerraformProvider{
		version: version,
		name:    terraform2.GOOGLE,
	}
	installer, err := terraform2.NewProviderInstaller(terraform2.ProviderConfig{
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
		Project: os.Getenv("CLOUDSDK_CORE_PROJECT"),
		Region:  os.Getenv("CLOUDSDK_COMPUTE_REGION"),
		Zone:    os.Getenv("CLOUDSDK_COMPUTE_ZONE"),
	}
}

func (p *GCPTerraformProvider) CheckCredentialsExist() error {
	client, err := asset.NewClient(context.Background())
	if err != nil {
		return errors.New("Please use a Service Account to authenticate on GCP.\n" +
			"For more information: https://cloud.google.com/docs/authentication/production")
	}
	_ = client.Close()
	return nil
}

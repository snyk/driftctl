package aws

import (
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

func InitTestAwsProvider(providerLibrary *terraform.ProviderLibrary) (*AWSTerraformProvider, error) {
	progress := &output.MockProgress{}
	progress.On("Inc").Maybe().Return()
	provider, err := NewAWSTerraformProvider("", progress)
	if err != nil {
		return nil, err
	}
	err = provider.Init()
	if err != nil {
		return nil, err
	}
	providerLibrary.AddProvider(terraform.AWS, provider)
	return provider, nil
}

package terraform

import (
	"github.com/sirupsen/logrus"
)

const (
	AWS    string = "aws"
	GITHUB string = "github"
	GOOGLE string = "google"
	AZURE  string = "azurerm"
)

type ProviderLibrary struct {
	providers map[string]TerraformProvider
}

func NewProviderLibrary() *ProviderLibrary {
	logrus.Debug("New provider library created")
	return &ProviderLibrary{
		make(map[string]TerraformProvider),
	}
}

func (p *ProviderLibrary) AddProvider(name string, provider TerraformProvider) {
	p.providers[name] = provider
}

func (p *ProviderLibrary) Provider(name string) TerraformProvider {
	return p.providers[name]
}

func (p *ProviderLibrary) Cleanup() {
	logrus.Debug("Closing providers")
	for providerKey, provider := range p.providers {
		logrus.WithFields(logrus.Fields{
			"key": providerKey,
		}).Debug("Closing provider")
		provider.Cleanup()
	}
}

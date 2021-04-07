package terraform

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	AWS    string = "aws"
	GITHUB string = "github"
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

func (p *ProviderLibrary) GetProviderForResourceType(resType string) (TerraformProvider, error) {

	var name string

	if strings.HasPrefix(resType, AWS) {
		name = AWS
	}

	if strings.HasPrefix(resType, GITHUB) {
		name = GITHUB
	}

	if name != "" {
		return p.Provider(name), nil
	}

	return nil, errors.New("Unable to resolve provider for resource")
}

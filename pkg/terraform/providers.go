package terraform

import (
	"github.com/hashicorp/go-plugin"
	"github.com/sirupsen/logrus"
)

const (
	AWS string = "aws"
)

var providers = make(map[string]TerraformProvider)

func AddProvider(name string, provider TerraformProvider) {
	providers[name] = provider
}

func Provider(name string) TerraformProvider {
	return providers[name]
}

func Providers() []TerraformProvider {
	m := make([]TerraformProvider, 0, len(providers))
	for _, val := range providers {
		m = append(m, val)
	}
	return m
}

func Cleanup() {
	logrus.Trace("Closing providers")
	plugin.CleanupClients()
}

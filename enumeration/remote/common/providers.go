package common

import (
	tf "github.com/snyk/driftctl/enumeration/terraform"
	"github.com/snyk/driftctl/enumeration/terraform/lock"
)

type RemoteParameter string

const (
	RemoteAWSTerraform        = "aws+tf"
	RemoteGithubTerraform     = "github+tf"
	RemoteGoogleTerraform     = "gcp+tf"
	RemoteGoogleBetaTerraform = "gcp-beta+tf"
	RemoteAzureTerraform      = "azure+tf"
)

var remoteParameterMapping = map[RemoteParameter]string{
	RemoteAWSTerraform:        tf.AWS,
	RemoteGithubTerraform:     tf.GITHUB,
	RemoteGoogleTerraform:     tf.GOOGLE,
	RemoteGoogleBetaTerraform: tf.GOOGLEBETA,
	RemoteAzureTerraform:      tf.AZURE,
}

func (p RemoteParameter) GetProviderAddress() *lock.ProviderAddress {
	return &lock.ProviderAddress{
		Hostname:  "registry.terraform.io",
		Namespace: "hashicorp",
		Type:      remoteParameterMapping[p],
	}
}

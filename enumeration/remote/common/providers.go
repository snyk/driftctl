package common

import (
	tf "github.com/snyk/driftctl/enumeration/terraform"
	"github.com/snyk/driftctl/enumeration/terraform/lock"
)

type RemoteParameter string

const (
	RemoteAWSTerraform      = "aws+tf"
	RemoteGithubTerraform   = "github+tf"
	RemoteGoogleTerraform   = "gcp+tf"
	RemoteAzureTerraform    = "azure+tf"
	RemoteScalewayTerraform = "scaleway+tf"
)

var remoteParameterMapping = map[RemoteParameter]string{
	RemoteAWSTerraform:      tf.AWS,
	RemoteGithubTerraform:   tf.GITHUB,
	RemoteGoogleTerraform:   tf.GOOGLE,
	RemoteAzureTerraform:    tf.AZURE,
	RemoteScalewayTerraform: tf.SCALEWAY,
}

func (p RemoteParameter) GetProviderAddress() *lock.ProviderAddress {
	return &lock.ProviderAddress{
		Hostname:  "registry.terraform.io",
		Namespace: "hashicorp",
		Type:      remoteParameterMapping[p],
	}
}

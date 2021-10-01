package common

import (
	tf "github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/pkg/terraform/lock"
)

type RemoteParameter string

const (
	RemoteAWSTerraform    = "aws+tf"
	RemoteGithubTerraform = "github+tf"
	RemoteGoogleTerraform = "gcp+tf"
	RemoteAzureTerraform  = "azure+tf"
)

var remoteParameterMapping = map[RemoteParameter]string{
	RemoteAWSTerraform:    tf.AWS,
	RemoteGithubTerraform: tf.GITHUB,
	RemoteGoogleTerraform: tf.GOOGLE,
	RemoteAzureTerraform:  tf.AZURE,
}

func (p RemoteParameter) GetProviderAddress() *lock.ProviderAddress {
	return &lock.ProviderAddress{
		Hostname:  "registry.terraform.io",
		Namespace: "hashicorp",
		Type:      remoteParameterMapping[p],
	}
}

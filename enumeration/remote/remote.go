package remote

import (
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/azurerm"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/remote/github"
	"github.com/snyk/driftctl/enumeration/remote/google"
	"github.com/snyk/driftctl/enumeration/remote/scaleway"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/terraform"
)

var supportedRemotes = []string{
	common.RemoteAWSTerraform,
	common.RemoteGithubTerraform,
	common.RemoteGoogleTerraform,
	common.RemoteAzureTerraform,
	common.RemoteScalewayTerraform,
}

func IsSupported(remote string) bool {
	for _, r := range supportedRemotes {
		if r == remote {
			return true
		}
	}
	return false
}

func Activate(remote, version string, alerter alerter.AlerterInterface, providerLibrary *terraform.ProviderLibrary, remoteLibrary *common.RemoteLibrary, progress enumeration.ProgressCounter, factory resource.ResourceFactory, configDir string) error {
	switch remote {
	case common.RemoteAWSTerraform:
		return aws.Init(version, alerter, providerLibrary, remoteLibrary, progress, factory, configDir)
	case common.RemoteGithubTerraform:
		return github.Init(version, alerter, providerLibrary, remoteLibrary, progress, factory, configDir)
	case common.RemoteGoogleTerraform:
		return google.Init(version, alerter, providerLibrary, remoteLibrary, progress, factory, configDir)
	case common.RemoteAzureTerraform:
		return azurerm.Init(version, alerter, providerLibrary, remoteLibrary, progress, factory, configDir)
	case common.RemoteScalewayTerraform:
		return scaleway.Init(version, alerter, providerLibrary, remoteLibrary, progress, factory, configDir)

	default:
		return errors.Errorf("unsupported remote '%s'", remote)
	}
}

func GetSupportedRemotes() []string {
	return supportedRemotes
}

package remote

import (
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/alerter"
	"github.com/snyk/driftctl/pkg/output"
	"github.com/snyk/driftctl/pkg/remote/aws"
	"github.com/snyk/driftctl/pkg/remote/azurerm"
	"github.com/snyk/driftctl/pkg/remote/common"
	"github.com/snyk/driftctl/pkg/remote/github"
	"github.com/snyk/driftctl/pkg/remote/google"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/terraform"
)

var supportedRemotes = []string{
	common.RemoteAWSTerraform,
	common.RemoteGithubTerraform,
	common.RemoteGoogleTerraform,
	common.RemoteAzureTerraform,
}

func IsSupported(remote string) bool {
	for _, r := range supportedRemotes {
		if r == remote {
			return true
		}
	}
	return false
}

func Activate(remote, version string, alerter *alerter.Alerter,
	providerLibrary *terraform.ProviderLibrary,
	remoteLibrary *common.RemoteLibrary,
	progress output.Progress,
	resourceSchemaRepository *resource.SchemaRepository,
	factory resource.ResourceFactory,
	configDir string) error {
	switch remote {
	case common.RemoteAWSTerraform:
		return aws.Init(version, alerter, providerLibrary, remoteLibrary, progress, resourceSchemaRepository, factory, configDir)
	case common.RemoteGithubTerraform:
		return github.Init(version, alerter, providerLibrary, remoteLibrary, progress, resourceSchemaRepository, factory, configDir)
	case common.RemoteGoogleTerraform:
		return google.Init(version, alerter, providerLibrary, remoteLibrary, progress, resourceSchemaRepository, factory, configDir)
	case common.RemoteAzureTerraform:
		return azurerm.Init(version, alerter, providerLibrary, remoteLibrary, progress, resourceSchemaRepository, factory, configDir)

	default:
		return errors.Errorf("unsupported remote '%s'", remote)
	}
}

func GetSupportedRemotes() []string {
	return supportedRemotes
}

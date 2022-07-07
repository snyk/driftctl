package resource

import (
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/snyk/driftctl/pkg/resource/azurerm"
	"github.com/snyk/driftctl/pkg/resource/github"
	"github.com/snyk/driftctl/pkg/resource/google"
)

func InitMetadatas(remote string,
	resourceSchemaRepository *resource.SchemaRepository) error {
	switch remote {
	case common.RemoteAWSTerraform:
		aws.InitResourcesMetadata(resourceSchemaRepository)
	case common.RemoteGithubTerraform:
		github.InitResourcesMetadata(resourceSchemaRepository)
	case common.RemoteGoogleTerraform:
		google.InitResourcesMetadata(resourceSchemaRepository)
	case common.RemoteAzureTerraform:
		azurerm.InitResourcesMetadata(resourceSchemaRepository)

	default:
		return errors.Errorf("unsupported remote '%s'", remote)
	}
	return nil
}

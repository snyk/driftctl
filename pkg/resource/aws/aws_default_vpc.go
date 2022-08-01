package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsDefaultVpcResourceType = "aws_default_vpc"

func initAwsDefaultVpcMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsDefaultVpcResourceType, resource.FlagDeepMode)
}

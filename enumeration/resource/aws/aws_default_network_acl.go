package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsDefaultNetworkACLResourceType = "aws_default_network_acl"

func initAwsDefaultNetworkACLMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsDefaultNetworkACLResourceType, resource.FlagDeepMode)
}

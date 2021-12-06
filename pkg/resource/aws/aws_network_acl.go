package aws

import (
	"github.com/snyk/driftctl/pkg/resource"
)

const AwsNetworkACLResourceType = "aws_network_acl"

func initAwsNetworkACLMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsNetworkACLResourceType, resource.FlagDeepMode)
}

package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsNetworkACLResourceType = "aws_network_acl"

func initAwsNetworkACLMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsNetworkACLResourceType, resource.FlagDeepMode)
}

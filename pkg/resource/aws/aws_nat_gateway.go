package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsNatGatewayResourceType = "aws_nat_gateway"

func initNatGatewayMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsNatGatewayResourceType, resource.FlagDeepMode)
}

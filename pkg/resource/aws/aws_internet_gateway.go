package aws

import (
	"github.com/snyk/driftctl/pkg/resource"
)

const AwsInternetGatewayResourceType = "aws_internet_gateway"

func initAwsInternetGatewayMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsInternetGatewayResourceType, resource.FlagDeepMode)
}

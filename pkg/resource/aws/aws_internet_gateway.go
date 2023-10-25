package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsInternetGatewayResourceType = "aws_internet_gateway"

func initAwsInternetGatewayMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsInternetGatewayResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"arn"})
	})
}

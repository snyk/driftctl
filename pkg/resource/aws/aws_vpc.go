package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsVpcResourceType = "aws_vpc"

func initAwsVpcMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsVpcResourceType, func(res *resource.Resource) {
		res.Attributes().SafeDelete([]string{"arn"})
	})
}

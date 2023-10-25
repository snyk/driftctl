package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsDefaultSubnetResourceType = "aws_default_subnet"

func initAwsDefaultSubnetMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsDefaultSubnetResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
}

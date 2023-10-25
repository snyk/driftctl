package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsDefaultRouteTableResourceType = "aws_default_route_table"

func initAwsDefaultRouteTableMetadata(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsDefaultRouteTableResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
}

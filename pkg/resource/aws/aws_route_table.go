package aws

import "github.com/snyk/driftctl/pkg/resource"

const AwsRouteTableResourceType = "aws_route_table"

func initAwsRouteTableMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsRouteTableResourceType, resource.FlagDeepMode)
	resourceSchemaRepository.SetNormalizeFunc(AwsRouteTableResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
}

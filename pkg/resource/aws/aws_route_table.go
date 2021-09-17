package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsRouteTableResourceType = "aws_route_table"

func initAwsRouteTableMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsRouteTableResourceType, resource.FlagDeepMode)
}

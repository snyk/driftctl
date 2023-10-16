package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsRouteTableAssociationResourceType = "aws_route_table_association"

func initAwsRouteTableAssociationMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsRouteTableAssociationResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if rtID := val.GetString("route_table_id"); rtID != nil && *rtID != "" {
			attrs["Table"] = *rtID
		}
		if gtwID := val.GetString("gateway_id"); gtwID != nil && *gtwID != "" {
			attrs["Gateway"] = *gtwID
		}
		if subnetID := val.GetString("subnet_id"); subnetID != nil && *subnetID != "" {
			attrs["Subnet"] = *subnetID
		}
		return attrs
	})
}

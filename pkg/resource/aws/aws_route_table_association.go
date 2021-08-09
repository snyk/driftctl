package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsRouteTableAssociationResourceType = "aws_route_table_association"

func initAwsRouteTableAssociationMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {

	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsRouteTableAssociationResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"route_table_id": *res.Attributes().GetString("route_table_id"),
		}
	})

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

package aws

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsRouteResourceType = "aws_route"

func initAwsRouteMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsRouteResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if rtID := val.GetString("route_table_id"); rtID != nil && *rtID != "" {
			attrs["Table"] = *rtID
		}
		if ipv4 := val.GetString("destination_cidr_block"); ipv4 != nil && *ipv4 != "" {
			attrs["Destination"] = *ipv4
		}
		if ipv6 := val.GetString("destination_ipv6_cidr_block"); ipv6 != nil && *ipv6 != "" {
			attrs["Destination"] = *ipv6
		}
		if prefix := val.GetString("destination_prefix_list_id"); prefix != nil && *prefix != "" {
			attrs["Destination"] = *prefix
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(AwsRouteResourceType, resource.FlagDeepMode)
}

func CalculateRouteID(tableId, CidrBlock, Ipv6CidrBlock, PrefixListId *string) string {
	if CidrBlock != nil && *CidrBlock != "" {
		return fmt.Sprintf("r-%s%d", *tableId, hashcode.String(*CidrBlock))
	}

	if Ipv6CidrBlock != nil && *Ipv6CidrBlock != "" {
		return fmt.Sprintf("r-%s%d", *tableId, hashcode.String(*Ipv6CidrBlock))
	}

	if PrefixListId != nil && *PrefixListId != "" {
		return fmt.Sprintf("r-%s%d", *tableId, hashcode.String(*PrefixListId))
	}

	return ""
}

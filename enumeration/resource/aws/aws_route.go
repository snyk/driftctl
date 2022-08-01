package aws

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/hashcode"
)

const AwsRouteResourceType = "aws_route"

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

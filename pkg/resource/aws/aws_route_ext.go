package aws

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/pkg/errors"
)

func (r *AwsRoute) String() string {
	var destination string
	if r.DestinationCidrBlock != nil && *r.DestinationCidrBlock != "" {
		destination = *r.DestinationCidrBlock
	}
	if r.DestinationIpv6CidrBlock != nil && *r.DestinationIpv6CidrBlock != "" {
		destination = *r.DestinationIpv6CidrBlock
	}
	output := fmt.Sprintf("Table: %s, Destination: %s", *r.RouteTableId, destination)
	return output
}

func CalculateRouteID(tableId, CidrBlock, Ipv6CidrBlock *string) (string, error) {
	if CidrBlock != nil && *CidrBlock != "" {
		return fmt.Sprintf("r-%s%d", *tableId, hashcode.String(*CidrBlock)), nil
	}

	if Ipv6CidrBlock != nil && *Ipv6CidrBlock != "" {
		return fmt.Sprintf("r-%s%d", *tableId, hashcode.String(*Ipv6CidrBlock)), nil
	}

	return "", errors.Errorf("invalid route detected for table %s", *tableId)
}

func (r *AwsRoute) NormalizeForState() (resource.Resource, error) {
	r.normalize()
	return r, nil
}

func (r *AwsRoute) NormalizeForProvider() (resource.Resource, error) {
	r.normalize()
	return r, nil
}

func (r *AwsRoute) normalize() {
	if r.VpcEndpointId != nil && *r.VpcEndpointId == "" {
		r.VpcEndpointId = nil
	}
	if r.LocalGatewayId != nil && *r.LocalGatewayId == "" {
		r.LocalGatewayId = nil
	}
	if r.DestinationIpv6CidrBlock != nil && *r.DestinationIpv6CidrBlock == "" {
		r.DestinationIpv6CidrBlock = nil
	}
	if r.DestinationCidrBlock != nil && *r.DestinationCidrBlock == "" {
		r.DestinationCidrBlock = nil
	}
	if r.EgressOnlyGatewayId != nil && *r.EgressOnlyGatewayId == "" {
		r.EgressOnlyGatewayId = nil
	}
	if r.InstanceId != nil && *r.InstanceId == "" {
		r.InstanceId = nil
	}
	if r.NatGatewayId != nil && *r.NatGatewayId == "" {
		r.NatGatewayId = nil
	}
	if r.NetworkInterfaceId != nil && *r.NetworkInterfaceId == "" {
		r.NetworkInterfaceId = nil
	}
	if r.TransitGatewayId != nil && *r.TransitGatewayId == "" {
		r.TransitGatewayId = nil
	}
	if r.VpcPeeringConnectionId != nil && *r.VpcPeeringConnectionId == "" {
		r.VpcPeeringConnectionId = nil
	}
}

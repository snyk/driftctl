package aws

import "fmt"

func (r *AwsRouteTableAssociation) String() string {
	assoc := fmt.Sprintf("Table: %s", *r.RouteTableId)
	if r.GatewayId != nil && *r.GatewayId != "" {
		assoc += fmt.Sprintf(", Gateway: %s", *r.GatewayId)
	}
	if r.SubnetId != nil && *r.SubnetId != "" {
		assoc += fmt.Sprintf(", Subnet: %s", *r.SubnetId)
	}
	return assoc
}

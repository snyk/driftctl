package aws

func (r *AwsRouteTableAssociation) Attributes() map[string]string {
	attrs := make(map[string]string)
	if r.RouteTableId != nil && *r.RouteTableId != "" {
		attrs["Table"] = *r.RouteTableId
	}
	if r.GatewayId != nil && *r.GatewayId != "" {
		attrs["Gateway"] = *r.GatewayId
	}
	if r.SubnetId != nil && *r.SubnetId != "" {
		attrs["Subnet"] = *r.SubnetId
	}
	return attrs
}

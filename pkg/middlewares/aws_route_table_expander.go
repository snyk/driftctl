package middlewares

import (
	"fmt"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type invalidRouteAlert struct {
	message string
}

func newInvalidRouteAlert(awsRouteTableResourceType, tableId string) *invalidRouteAlert {
	message := fmt.Sprintf("Skipped invalid route found in state for %s.%s", awsRouteTableResourceType, tableId)
	return &invalidRouteAlert{message}
}

func (i *invalidRouteAlert) Message() string {
	return i.message
}

func (i *invalidRouteAlert) ShouldIgnoreResource() bool {
	return false
}

// Explodes routes found in aws_default_route_table.route and aws_route_table.route to dedicated resources
type AwsRouteTableExpander struct {
	alerter         alerter.AlerterInterface
	resourceFactory resource.ResourceFactory
}

func NewAwsRouteTableExpander(alerter alerter.AlerterInterface, resourceFactory resource.ResourceFactory) AwsRouteTableExpander {
	return AwsRouteTableExpander{
		alerter,
		resourceFactory,
	}
}

func (m AwsRouteTableExpander) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {

	newList := make([]resource.Resource, 0, len(*resourcesFromState))
	for _, res := range *resourcesFromState {

		// Ignore all resources other than (default) routes tables
		if res.TerraformType() != aws.AwsRouteTableResourceType &&
			res.TerraformType() != aws.AwsDefaultRouteTableResourceType {
			newList = append(newList, res)
			continue
		}

		table, _ := res.(*aws.AwsRouteTable)
		defaultTable, isDefault := res.(*aws.AwsDefaultRouteTable)
		newList = append(newList, res)

		var err error
		if isDefault {
			err = m.handleDefaultTable(defaultTable, &newList, *resourcesFromState)
		} else {
			err = m.handleTable(table, &newList, *resourcesFromState)
		}

		if err != nil {
			return err
		}
	}
	*resourcesFromState = newList
	return nil
}

func (m *AwsRouteTableExpander) handleTable(table *aws.AwsRouteTable, results *[]resource.Resource, resourcesFromState []resource.Resource) error {
	if table.Route == nil ||
		len(*table.Route) < 1 {
		return nil
	}
	for _, route := range *table.Route {
		routeId, err := aws.CalculateRouteID(&table.Id, route.CidrBlock, route.Ipv6CidrBlock)
		if err != nil {
			m.alerter.SendAlert(aws.AwsRouteTableResourceType, newInvalidRouteAlert(aws.AwsRouteTableResourceType, table.Id))
			continue
		}

		data := map[string]interface{}{
			"destination_cidr_block":      route.CidrBlock,
			"destination_ipv6_cidr_block": route.Ipv6CidrBlock,
			"destination_prefix_list_id":  "",
			"egress_only_gateway_id":      route.EgressOnlyGatewayId,
			"gateway_id":                  route.GatewayId,
			"id":                          routeId,
			"instance_id":                 route.InstanceId,
			"instance_owner_id":           "",
			"local_gateway_id":            route.LocalGatewayId,
			"nat_gateway_id":              route.NatGatewayId,
			"network_interface_id":        route.NetworkInterfaceId,
			"origin":                      "CreateRoute",
			"route_table_id":              table.Id,
			"state":                       "active",
			"transit_gateway_id":          route.TransitGatewayId,
			"vpc_endpoint_id":             route.VpcEndpointId,
			"vpc_peering_connection_id":   route.VpcPeeringConnectionId,
		}
		ctyVal, err := m.resourceFactory.CreateResource(data, "aws_route")
		if err != nil {
			return err
		}

		// Don't expand if the route already exists as a dedicated resource
		if m.routeExists(routeId, resourcesFromState) {
			continue
		}

		newRouteFromTable := &aws.AwsRoute{
			DestinationCidrBlock:     route.CidrBlock,
			DestinationIpv6CidrBlock: route.Ipv6CidrBlock,
			DestinationPrefixListId:  awssdk.String(""),
			EgressOnlyGatewayId:      route.EgressOnlyGatewayId,
			GatewayId:                route.GatewayId,
			Id:                       routeId,
			InstanceId:               route.InstanceId,
			InstanceOwnerId:          awssdk.String(""),
			LocalGatewayId:           route.LocalGatewayId,
			NatGatewayId:             route.NatGatewayId,
			NetworkInterfaceId:       route.NetworkInterfaceId,
			Origin:                   awssdk.String("CreateRoute"),
			RouteTableId:             awssdk.String(table.Id),
			State:                    awssdk.String("active"),
			TransitGatewayId:         route.TransitGatewayId,
			VpcEndpointId:            route.VpcEndpointId,
			VpcPeeringConnectionId:   route.VpcPeeringConnectionId,
			CtyVal:                   ctyVal,
		}
		normalizedRes, err := newRouteFromTable.NormalizeForState()
		if err != nil {
			return err
		}
		*results = append(*results, normalizedRes)
		logrus.WithFields(logrus.Fields{
			"route": newRouteFromTable.String(),
		}).Debug("Created new route from route table")
	}

	table.Route = nil

	return nil
}

func (m *AwsRouteTableExpander) handleDefaultTable(table *aws.AwsDefaultRouteTable, results *[]resource.Resource, resourcesFromState []resource.Resource) error {
	if table.Route == nil ||
		len(*table.Route) < 1 {
		return nil
	}
	for _, route := range *table.Route {
		routeId, err := aws.CalculateRouteID(&table.Id, route.CidrBlock, route.Ipv6CidrBlock)
		if err != nil {
			m.alerter.SendAlert(aws.AwsDefaultRouteTableResourceType, newInvalidRouteAlert(aws.AwsDefaultRouteTableResourceType, table.Id))
			continue
		}

		data := map[string]interface{}{
			"destination_cidr_block":      route.CidrBlock,
			"destination_ipv6_cidr_block": route.Ipv6CidrBlock,
			"destination_prefix_list_id":  "",
			"egress_only_gateway_id":      route.EgressOnlyGatewayId,
			"gateway_id":                  route.GatewayId,
			"id":                          routeId,
			"instance_id":                 route.InstanceId,
			"instance_owner_id":           "",
			"nat_gateway_id":              route.NatGatewayId,
			"network_interface_id":        route.NetworkInterfaceId,
			"origin":                      "CreateRoute",
			"route_table_id":              table.Id,
			"state":                       "active",
			"transit_gateway_id":          route.TransitGatewayId,
			"vpc_endpoint_id":             route.VpcEndpointId,
			"vpc_peering_connection_id":   route.VpcPeeringConnectionId,
		}
		ctyVal, err := m.resourceFactory.CreateResource(data, "aws_route")
		if err != nil {
			return err
		}

		// Don't expand if the route already exists as a dedicated resource
		if m.routeExists(routeId, resourcesFromState) {
			continue
		}

		newRouteFromTable := &aws.AwsRoute{
			DestinationCidrBlock:     route.CidrBlock,
			DestinationIpv6CidrBlock: route.Ipv6CidrBlock,
			DestinationPrefixListId:  awssdk.String(""),
			EgressOnlyGatewayId:      route.EgressOnlyGatewayId,
			GatewayId:                route.GatewayId,
			Id:                       routeId,
			InstanceId:               route.InstanceId,
			InstanceOwnerId:          awssdk.String(""),
			NatGatewayId:             route.NatGatewayId,
			NetworkInterfaceId:       route.NetworkInterfaceId,
			Origin:                   awssdk.String("CreateRoute"),
			RouteTableId:             awssdk.String(table.Id),
			State:                    awssdk.String("active"),
			TransitGatewayId:         route.TransitGatewayId,
			VpcEndpointId:            route.VpcEndpointId,
			VpcPeeringConnectionId:   route.VpcPeeringConnectionId,
			CtyVal:                   ctyVal,
		}
		normalizedRes, err := newRouteFromTable.NormalizeForState()
		if err != nil {
			return err
		}
		*results = append(*results, normalizedRes)
		logrus.WithFields(logrus.Fields{
			"route": newRouteFromTable.String(),
		}).Debug("Created new route from default route table")
	}

	table.Route = nil

	return nil
}

func (m *AwsRouteTableExpander) routeExists(routeId string, resourcesFromState []resource.Resource) bool {
	for _, res := range resourcesFromState {
		if res.TerraformType() == aws.AwsRouteResourceType && res.TerraformId() == routeId {
			return true
		}
	}

	return false
}

package middlewares

import (
	"fmt"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// Explodes routes found in aws_default_route_table.route and aws_route_table.route to dedicated resources
type AwsRouteTableExpander struct {
	alerter alerter.AlerterInterface
}

func NewAwsRouteTableExpander(alerter alerter.AlerterInterface) AwsRouteTableExpander {
	return AwsRouteTableExpander{
		alerter,
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
			err = m.handleDefaultTable(defaultTable, &newList)
		} else {
			err = m.handleTable(table, &newList)
		}

		if err != nil {
			return err
		}
	}
	*resourcesFromState = newList
	return nil
}

func (m *AwsRouteTableExpander) handleTable(table *aws.AwsRouteTable, results *[]resource.Resource) error {
	if table.Route == nil ||
		len(*table.Route) < 1 {
		return nil
	}
	for _, route := range *table.Route {
		routeId, err := aws.CalculateRouteID(&table.Id, route.CidrBlock, route.Ipv6CidrBlock)
		if err != nil {
			m.alerter.SendAlert(aws.AwsRouteTableResourceType, alerter.Alert{
				Message: fmt.Sprintf("Skipped invalid route found in state for %s.%s", aws.AwsRouteTableResourceType, table.Id),
			})
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

func (m *AwsRouteTableExpander) handleDefaultTable(table *aws.AwsDefaultRouteTable, results *[]resource.Resource) error {
	if table.Route == nil ||
		len(*table.Route) < 1 {
		return nil
	}
	for _, route := range *table.Route {
		routeId, err := aws.CalculateRouteID(&table.Id, route.CidrBlock, route.Ipv6CidrBlock)
		if err != nil {
			m.alerter.SendAlert(aws.AwsDefaultRouteTableResourceType, alerter.Alert{
				Message: fmt.Sprintf("Skipped invalid route found in state for %s.%s", aws.AwsDefaultRouteTableResourceType, table.Id),
			})
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

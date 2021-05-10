package middlewares

import (
	"fmt"

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

		table, _ := res.(*resource.AbstractResource)
		newList = append(newList, res)

		var err error
		if res.TerraformType() == aws.AwsDefaultRouteTableResourceType {
			err = m.handleDefaultTable(table, &newList, *resourcesFromState)
		} else {
			err = m.handleTable(table, &newList, *resourcesFromState)
		}

		if err != nil {
			return err
		}
	}

	newRemoteResources := make([]resource.Resource, 0)
	for _, remoteRes := range *remoteResources {
		if remoteRes.TerraformType() != aws.AwsRouteTableResourceType &&
			remoteRes.TerraformType() != aws.AwsDefaultRouteTableResourceType {
			newRemoteResources = append(newRemoteResources, remoteRes)
			continue
		}
		table, _ := remoteRes.(*resource.AbstractResource)
		table.Attrs.SafeDelete([]string{"route"})
		newRemoteResources = append(newRemoteResources, table)
	}

	*resourcesFromState = newList
	*remoteResources = newRemoteResources
	return nil
}

func (m *AwsRouteTableExpander) handleTable(table *resource.AbstractResource, results *[]resource.Resource, resourcesFromState []resource.Resource) error {
	routes, exist := table.Attrs.Get("route")
	if !exist || routes == nil {
		return nil
	}
	for _, route := range routes.([]interface{}) {
		route := route.(map[string]interface{})
		cidrBlock := ""
		if route["cidr_block"] != nil {
			cidrBlock = route["cidr_block"].(string)
		}
		ipv6CidrBlock := ""
		if route["ipv6_cidr_block"] != nil {
			ipv6CidrBlock = route["ipv6_cidr_block"].(string)
		}
		routeId, err := aws.CalculateRouteID(&table.Id, &cidrBlock, &ipv6CidrBlock)
		if err != nil {
			m.alerter.SendAlert(aws.AwsRouteTableResourceType, newInvalidRouteAlert(aws.AwsRouteTableResourceType, table.Id))
			continue
		}

		data := map[string]interface{}{
			"destination_cidr_block":      route["cidr_block"],
			"destination_ipv6_cidr_block": route["ipv6_cidr_block"],
			"egress_only_gateway_id":      route["egress_only_gateway_id"],
			"gateway_id":                  route["gateway_id"],
			"id":                          routeId,
			"instance_id":                 route["instance_id"],
			"instance_owner_id":           "",
			"local_gateway_id":            route["local_gateway_id"],
			"nat_gateway_id":              route["nat_gateway_id"],
			"network_interface_id":        route["network_interface_id"],
			"origin":                      "CreateRoute",
			"route_table_id":              table.Id,
			"state":                       "active",
			"transit_gateway_id":          route["transit_gateway_id"],
			"vpc_endpoint_id":             route["vpc_endpoint_id"],
			"vpc_peering_connection_id":   route["vpc_peering_connection_id"],
		}
		// Don't expand if the route already exists as a dedicated resource
		if m.routeExists(routeId, resourcesFromState) {
			continue
		}
		newRes := m.resourceFactory.CreateAbstractResource(aws.AwsRouteResourceType, routeId, data)
		*results = append(*results, newRes)
		logrus.WithFields(logrus.Fields{
			"route": routeId,
		}).Debug("Created new route from route table")
	}
	table.Attrs.SafeDelete([]string{"route"})
	return nil
}

func (m *AwsRouteTableExpander) handleDefaultTable(table *resource.AbstractResource, results *[]resource.Resource, resourcesFromState []resource.Resource) error {
	routes, exist := table.Attrs.Get("route")
	if !exist || routes == nil {
		return nil
	}
	for _, route := range routes.([]interface{}) {
		route := route.(map[string]interface{})
		cidrBlock := ""
		if route["cidr_block"] != nil {
			cidrBlock = route["cidr_block"].(string)
		}
		ipv6CidrBlock := ""
		if route["ipv6_cidr_block"] != nil {
			ipv6CidrBlock = route["ipv6_cidr_block"].(string)
		}
		routeId, err := aws.CalculateRouteID(&table.Id, &cidrBlock, &ipv6CidrBlock)
		if err != nil {
			m.alerter.SendAlert(aws.AwsDefaultRouteTableResourceType, newInvalidRouteAlert(aws.AwsDefaultRouteTableResourceType, table.Id))
			continue
		}

		data := map[string]interface{}{
			"destination_cidr_block":      route["cidr_block"],
			"destination_ipv6_cidr_block": route["ipv6_cidr_block"],
			"egress_only_gateway_id":      route["egress_only_gateway_id"],
			"gateway_id":                  route["gateway_id"],
			"id":                          routeId,
			"instance_id":                 route["instance_id"],
			"nat_gateway_id":              route["nat_gateway_id"],
			"network_interface_id":        route["network_interface_id"],
			"origin":                      "CreateRoute",
			"route_table_id":              table.Id,
			"state":                       "active",
			"transit_gateway_id":          route["transit_gateway_id"],
			"vpc_endpoint_id":             route["vpc_endpoint_id"],
			"vpc_peering_connection_id":   route["vpc_peering_connection_id"],
		}
		// Don't expand if the route already exists as a dedicated resource
		if m.routeExists(routeId, resourcesFromState) {
			continue
		}
		newRes := m.resourceFactory.CreateAbstractResource(aws.AwsRouteResourceType, routeId, data)
		*results = append(*results, newRes)
		logrus.WithFields(logrus.Fields{
			"route": routeId,
		}).Debug("Created new route from default route table")
	}
	table.Attrs.SafeDelete([]string{"route"})
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

package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/alerter"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

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

func (m AwsRouteTableExpander) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newList := make([]*resource.Resource, 0, len(*resourcesFromState))
	for _, res := range *resourcesFromState {

		// Ignore all resources other than (default) routes tables
		if res.ResourceType() != aws.AwsRouteTableResourceType &&
			res.ResourceType() != aws.AwsDefaultRouteTableResourceType {
			newList = append(newList, res)
			continue
		}

		newList = append(newList, res)

		var err error
		if res.ResourceType() == aws.AwsDefaultRouteTableResourceType {
			err = m.handleDefaultTable(res, &newList, *resourcesFromState)
		} else {
			err = m.handleTable(res, &newList, *resourcesFromState)
		}

		if err != nil {
			return err
		}
	}

	newRemoteResources := make([]*resource.Resource, 0)
	for _, remoteRes := range *remoteResources {
		if remoteRes.ResourceType() != aws.AwsRouteTableResourceType &&
			remoteRes.ResourceType() != aws.AwsDefaultRouteTableResourceType {
			newRemoteResources = append(newRemoteResources, remoteRes)
			continue
		}
		remoteRes.Attrs.SafeDelete([]string{"route"})
		newRemoteResources = append(newRemoteResources, remoteRes)
	}

	*resourcesFromState = newList
	*remoteResources = newRemoteResources
	return nil
}

func (m *AwsRouteTableExpander) handleTable(table *resource.Resource, results *[]*resource.Resource, resourcesFromState []*resource.Resource) error {
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
		prefixListId := ""
		if route["destination_prefix_list_id"] != nil {
			prefixListId = route["destination_prefix_list_id"].(string)
		}
		routeId := aws.CalculateRouteID(&table.Id, &cidrBlock, &ipv6CidrBlock, &prefixListId)

		data := map[string]interface{}{
			"destination_cidr_block":      route["cidr_block"],
			"destination_ipv6_cidr_block": route["ipv6_cidr_block"],
			"destination_prefix_list_id":  route["destination_prefix_list_id"],
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

func (m *AwsRouteTableExpander) handleDefaultTable(table *resource.Resource, results *[]*resource.Resource, resourcesFromState []*resource.Resource) error {
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
		prefixListId := ""
		if route["destination_prefix_list_id"] != nil {
			prefixListId = route["destination_prefix_list_id"].(string)
		}
		routeId := aws.CalculateRouteID(&table.Id, &cidrBlock, &ipv6CidrBlock, &prefixListId)

		data := map[string]interface{}{
			"destination_cidr_block":      route["cidr_block"],
			"destination_ipv6_cidr_block": route["ipv6_cidr_block"],
			"destination_prefix_list_id":  route["destination_prefix_list_id"],
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

func (m *AwsRouteTableExpander) routeExists(routeId string, resourcesFromState []*resource.Resource) bool {
	for _, res := range resourcesFromState {
		if res.ResourceType() == aws.AwsRouteResourceType && res.ResourceId() == routeId {
			return true
		}
	}

	return false
}

package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/hashicorp/terraform/flatmap"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type RouteSupplier struct {
	reader            terraform.ResourceReader
	routeDeserializer deserializer.CTYDeserializer
	client            ec2iface.EC2API
	routeRunner       *terraform.ParallelResourceReader
}

func NewRouteSupplier(runner *parallel.ParallelRunner, client ec2iface.EC2API) *RouteSupplier {
	return &RouteSupplier{
		terraform.Provider(terraform.AWS),
		awsdeserializer.NewRouteDeserializer(),
		client,
		terraform.NewParallelResourceReader(runner.SubRunner()),
	}
}

func (s RouteSupplier) Resources() ([]resource.Resource, error) {

	routeTables, err := listRouteTables(s.client, aws.AwsRouteResourceType)
	if err != nil {
		return nil, err
	}

	for _, routeTable := range routeTables {
		table := *routeTable
		for _, route := range table.Routes {
			res := *route
			s.routeRunner.Run(func() (cty.Value, error) {
				return s.readRoute(*table.RouteTableId, res)
			})
		}
	}

	routeResources, err := s.routeRunner.Wait()
	if err != nil {
		return nil, err
	}

	deserializedRoutes, err := s.routeDeserializer.Deserialize(routeResources)
	if err != nil {
		return nil, err
	}

	return deserializedRoutes, nil
}

func (s RouteSupplier) readRoute(tableId string, route ec2.Route) (cty.Value, error) {
	var Ty resource.ResourceType = aws.AwsRouteResourceType

	attributes := map[string]interface{}{
		"route_table_id": tableId,
	}
	if route.DestinationCidrBlock != nil && *route.DestinationCidrBlock != "" {
		attributes["destination_cidr_block"] = *route.DestinationCidrBlock
	}

	if route.DestinationIpv6CidrBlock != nil && *route.DestinationIpv6CidrBlock != "" {
		attributes["destination_ipv6_cidr_block"] = *route.DestinationIpv6CidrBlock
	}

	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID:         aws.CalculateRouteID(&tableId, route.DestinationCidrBlock, route.DestinationIpv6CidrBlock),
		Ty:         Ty,
		Attributes: flatmap.Flatten(attributes),
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type": Ty,
		}).Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}

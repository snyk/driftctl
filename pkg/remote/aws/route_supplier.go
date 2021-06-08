package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/hashicorp/terraform/flatmap"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type RouteSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repo         repository.EC2Repository
	routeRunner  *terraform.ParallelResourceReader
}

func NewRouteSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.EC2Repository) *RouteSupplier {
	return &RouteSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *RouteSupplier) SuppliedType() resource.ResourceType {
	return aws.AwsRouteResourceType
}

func (s *RouteSupplier) Resources() ([]resource.Resource, error) {
	routeTables, err := s.repo.ListAllRouteTables()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, s.SuppliedType(), aws.AwsRouteTableResourceType)
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

	deserializedRoutes, err := s.deserializer.Deserialize(s.SuppliedType(), routeResources)
	if err != nil {
		return nil, err
	}

	return deserializedRoutes, nil
}

func (s *RouteSupplier) readRoute(tableId string, route ec2.Route) (cty.Value, error) {
	var Ty resource.ResourceType = s.SuppliedType()

	attributes := map[string]interface{}{
		"route_table_id": tableId,
	}
	if route.DestinationCidrBlock != nil && *route.DestinationCidrBlock != "" {
		attributes["destination_cidr_block"] = *route.DestinationCidrBlock
	}

	if route.DestinationIpv6CidrBlock != nil && *route.DestinationIpv6CidrBlock != "" {
		attributes["destination_ipv6_cidr_block"] = *route.DestinationIpv6CidrBlock
	}

	// We can ignore error there as remote will always return us a valid route
	routeId, _ := aws.CalculateRouteID(&tableId, route.DestinationCidrBlock, route.DestinationIpv6CidrBlock)
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID:         routeId,
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

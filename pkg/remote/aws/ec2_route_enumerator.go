package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type EC2RouteEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2RouteEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2RouteEnumerator {
	return &EC2RouteEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2RouteEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsRouteResourceType
}

func (e *EC2RouteEnumerator) Enumerate() ([]resource.Resource, error) {
	routeTables, err := e.repository.ListAllRouteTables()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, string(e.SupportedType()), aws.AwsRouteTableResourceType)
	}

	var results []resource.Resource

	for _, routeTable := range routeTables {
		for _, route := range routeTable.Routes {
			routeId, _ := aws.CalculateRouteID(routeTable.RouteTableId, route.DestinationCidrBlock, route.DestinationIpv6CidrBlock)
			data := map[string]interface{}{
				"route_table_id": *routeTable.RouteTableId,
				"origin":         *route.Origin,
			}
			if route.DestinationCidrBlock != nil && *route.DestinationCidrBlock != "" {
				data["destination_cidr_block"] = *route.DestinationCidrBlock
			}
			if route.DestinationIpv6CidrBlock != nil && *route.DestinationIpv6CidrBlock != "" {
				data["destination_ipv6_cidr_block"] = *route.DestinationIpv6CidrBlock
			}
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					routeId,
					data,
				),
			)
		}
	}

	return results, err
}

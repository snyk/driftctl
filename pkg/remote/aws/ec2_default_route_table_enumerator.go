package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type EC2DefaultRouteTableEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2DefaultRouteTableEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2DefaultRouteTableEnumerator {
	return &EC2DefaultRouteTableEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2DefaultRouteTableEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsDefaultRouteTableResourceType
}

func (e *EC2DefaultRouteTableEnumerator) Enumerate() ([]*resource.Resource, error) {
	routeTables, err := e.repository.ListAllRouteTables()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	var results []*resource.Resource

	for _, routeTable := range routeTables {
		if isMainRouteTable(routeTable) {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*routeTable.RouteTableId,
					map[string]interface{}{
						"vpc_id": *routeTable.VpcId,
					},
				),
			)
		}
	}

	return results, err
}

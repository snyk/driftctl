package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type EC2RouteTableEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2RouteTableEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2RouteTableEnumerator {
	return &EC2RouteTableEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2RouteTableEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsRouteTableResourceType
}

func (e *EC2RouteTableEnumerator) Enumerate() ([]*resource.Resource, error) {
	routeTables, err := e.repository.ListAllRouteTables()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	var results []*resource.Resource

	for _, routeTable := range routeTables {
		if !isMainRouteTable(routeTable) {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*routeTable.RouteTableId,
					map[string]interface{}{},
				),
			)
		}
	}

	return results, err
}

func isMainRouteTable(routeTable *ec2.RouteTable) bool {
	for _, assoc := range routeTable.Associations {
		if assoc.Main != nil && *assoc.Main {
			return true
		}
	}
	return false
}

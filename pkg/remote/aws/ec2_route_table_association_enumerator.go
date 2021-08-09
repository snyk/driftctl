package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type EC2RouteTableAssociationEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2RouteTableAssociationEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2RouteTableAssociationEnumerator {
	return &EC2RouteTableAssociationEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2RouteTableAssociationEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsRouteTableAssociationResourceType
}

func (e *EC2RouteTableAssociationEnumerator) Enumerate() ([]*resource.Resource, error) {
	routeTables, err := e.repository.ListAllRouteTables()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsRouteTableResourceType)
	}

	var results []*resource.Resource

	for _, routeTable := range routeTables {
		for _, assoc := range routeTable.Associations {
			if e.shouldBeIgnored(assoc) {
				continue
			}
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*assoc.RouteTableAssociationId,
					map[string]interface{}{
						"route_table_id": *assoc.RouteTableId,
					},
				),
			)
		}
	}

	return results, err
}

func (e *EC2RouteTableAssociationEnumerator) shouldBeIgnored(assoc *ec2.RouteTableAssociation) bool {
	// Ignore when nothing is associated
	if assoc.GatewayId == nil && assoc.SubnetId == nil {
		return true
	}

	// Ignore when association is not associated
	if assoc.AssociationState != nil && assoc.AssociationState.State != nil &&
		*assoc.AssociationState.State != "associated" {
		return true
	}

	return false
}

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

type RouteTableAssociationSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repo         repository.EC2Repository
	runner       *terraform.ParallelResourceReader
}

func NewRouteTableAssociationSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.EC2Repository) *RouteTableAssociationSupplier {
	return &RouteTableAssociationSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *RouteTableAssociationSupplier) SuppliedType() resource.ResourceType {
	return aws.AwsRouteTableAssociationResourceType
}

func (s *RouteTableAssociationSupplier) Resources() ([]resource.Resource, error) {
	tables, err := s.repo.ListAllRouteTables()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, s.SuppliedType(), aws.AwsRouteTableResourceType)
	}

	for _, t := range tables {
		table := *t
		for _, assoc := range table.Associations {
			res := *assoc
			if s.shouldBeIgnored(assoc) {
				continue
			}
			s.runner.Run(func() (cty.Value, error) {
				return s.readRouteTableAssociation(res)
			})
		}
	}

	// Retrieve results from terraform provider
	routeTableAssociationResources, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	// Deserialize
	deserializedRouteTableAssociations, err := s.deserializer.Deserialize(s.SuppliedType(), routeTableAssociationResources)
	if err != nil {
		return nil, err
	}

	return deserializedRouteTableAssociations, nil
}

func (s *RouteTableAssociationSupplier) readRouteTableAssociation(assoc ec2.RouteTableAssociation) (cty.Value, error) {
	var Ty resource.ResourceType = s.SuppliedType()
	attributes := map[string]interface{}{
		"route_table_id": *assoc.RouteTableId,
	}
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID:         *assoc.RouteTableAssociationId,
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

func (s *RouteTableAssociationSupplier) shouldBeIgnored(assoc *ec2.RouteTableAssociation) bool {

	// Ignore when nothing is associated
	if assoc.GatewayId == nil && assoc.SubnetId == nil {
		return true
	}

	// Ignore when association not associated
	if assoc.AssociationState != nil && assoc.AssociationState.State != nil &&
		*assoc.AssociationState.State != "associated" {
		return true
	}

	return false
}

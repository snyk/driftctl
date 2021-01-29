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

type RouteTableAssociationSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       ec2iface.EC2API
	runner       *terraform.ParallelResourceReader
}

func NewRouteTableAssociationSupplier(runner *parallel.ParallelRunner, client ec2iface.EC2API) *RouteTableAssociationSupplier {
	return &RouteTableAssociationSupplier{
		terraform.Provider(terraform.AWS),
		awsdeserializer.NewRouteTableAssociationDeserializer(),
		client,
		terraform.NewParallelResourceReader(runner),
	}
}

func (s RouteTableAssociationSupplier) Resources() ([]resource.Resource, error) {

	tables, err := listRouteTables(s.client, aws.AwsRouteTableAssociationResourceType)
	if err != nil {
		return nil, err
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
	deserializedRouteTableAssociations, err := s.deserializer.Deserialize(routeTableAssociationResources)
	if err != nil {
		return nil, err
	}

	return deserializedRouteTableAssociations, nil
}

func (s RouteTableAssociationSupplier) readRouteTableAssociation(assoc ec2.RouteTableAssociation) (cty.Value, error) {
	var Ty resource.ResourceType = aws.AwsRouteTableAssociationResourceType
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

func (s RouteTableAssociationSupplier) shouldBeIgnored(assoc *ec2.RouteTableAssociation) bool {

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

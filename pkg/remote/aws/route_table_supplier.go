package aws

import (
	"errors"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/hashicorp/terraform/flatmap"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type RouteTableSupplier struct {
	reader                        terraform.ResourceReader
	defaultRouteTableDeserializer deserializer.CTYDeserializer
	routeTableDeserializer        deserializer.CTYDeserializer
	client                        repository.EC2Repository
	defaultRouteTableRunner       *terraform.ParallelResourceReader
	routeTableRunner              *terraform.ParallelResourceReader
}

func NewRouteTableSupplier(provider *AWSTerraformProvider) *RouteTableSupplier {
	return &RouteTableSupplier{
		provider,
		awsdeserializer.NewDefaultRouteTableDeserializer(),
		awsdeserializer.NewRouteTableDeserializer(),
		repository.NewEC2Repository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *RouteTableSupplier) Resources() ([]resource.Resource, error) {
	results, err := s.client.ListAllRouteTables()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsRouteTableResourceType)
	}

	retrievedDefaultRouteTables := []*ec2.RouteTable{}
	retrievedRouteTables := []*ec2.RouteTable{}

	for _, routeTable := range results {
		res := *routeTable
		var isMain bool
		for _, assoc := range res.Associations {
			if assoc.Main != nil && *assoc.Main {
				isMain = true
				break
			}
		}
		if isMain {
			retrievedDefaultRouteTables = append(retrievedDefaultRouteTables, &res)
			continue
		}
		retrievedRouteTables = append(retrievedRouteTables, &res)
	}

	for _, routeTable := range retrievedDefaultRouteTables {
		res := *routeTable
		s.defaultRouteTableRunner.Run(func() (cty.Value, error) {
			return s.readRouteTable(res, true)
		})
	}

	// Retrieve results from terraform provider
	defaultRouteTableResources, err := s.defaultRouteTableRunner.Wait()
	if err != nil {
		return nil, err
	}

	for _, routeTable := range retrievedRouteTables {
		res := *routeTable
		s.routeTableRunner.Run(func() (cty.Value, error) {
			return s.readRouteTable(res, false)
		})
	}

	routeTableResources, err := s.routeTableRunner.Wait()
	if err != nil {
		return nil, err
	}

	// Deserialize
	deserializedDefaultRouteTables, err := s.defaultRouteTableDeserializer.Deserialize(defaultRouteTableResources)
	if err != nil {
		return nil, err
	}
	deserializedRouteTables, err := s.routeTableDeserializer.Deserialize(routeTableResources)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0, len(routeTableResources)+len(defaultRouteTableResources))
	resources = append(resources, deserializedDefaultRouteTables...)
	resources = append(resources, deserializedRouteTables...)

	return resources, nil
}

func (s *RouteTableSupplier) readRouteTable(routeTable ec2.RouteTable, isMain bool) (cty.Value, error) {
	var Ty resource.ResourceType = aws.AwsRouteTableResourceType
	attributes := map[string]interface{}{}
	if isMain {
		if routeTable.VpcId == nil {
			return cty.NilVal, errors.New("a default route table does not have a VpcId")
		}
		Ty = aws.AwsDefaultRouteTableResourceType
		attributes["vpc_id"] = *routeTable.VpcId
	}
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID:         *routeTable.RouteTableId,
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

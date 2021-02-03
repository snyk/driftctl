package aws

import (
	"errors"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
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
	client                        ec2iface.EC2API
	defaultRouteTableRunner       *terraform.ParallelResourceReader
	routeTableRunner              *terraform.ParallelResourceReader
}

func NewRouteTableSupplier(provider *TerraformProvider) *RouteTableSupplier {
	return &RouteTableSupplier{
		provider,
		awsdeserializer.NewDefaultRouteTableDeserializer(),
		awsdeserializer.NewRouteTableDeserializer(),
		ec2.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s RouteTableSupplier) Resources() ([]resource.Resource, error) {

	retrievedRouteTables, err := listRouteTables(s.client, aws.AwsRouteTableResourceType)
	if err != nil {
		return nil, err
	}

	for _, routeTable := range retrievedRouteTables {
		res := *routeTable
		var isMain bool
		for _, assoc := range res.Associations {
			if assoc.Main != nil && *assoc.Main {
				isMain = true
				break
			}
		}
		if isMain {
			s.defaultRouteTableRunner.Run(func() (cty.Value, error) {
				return s.readRouteTable(res, true)
			})
			continue
		}
		s.routeTableRunner.Run(func() (cty.Value, error) {
			return s.readRouteTable(res, false)
		})
	}

	// Retrieve results from terraform provider
	defaultRouteTableResources, err := s.defaultRouteTableRunner.Wait()
	if err != nil {
		return nil, err
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

func (s RouteTableSupplier) readRouteTable(routeTable ec2.RouteTable, isMain bool) (cty.Value, error) {
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

func listRouteTables(client ec2iface.EC2API, supplierType string) ([]*ec2.RouteTable, error) {
	var routeTables []*ec2.RouteTable
	input := ec2.DescribeRouteTablesInput{}
	err := client.DescribeRouteTablesPages(&input,
		func(resp *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
			routeTables = append(routeTables, resp.RouteTables...)
			return !lastPage
		},
	)

	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, supplierType, aws.AwsRouteTableResourceType)
	}

	return routeTables, nil
}

package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws"
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRouteTableAssociationSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockEC2Repository)
		err     error
	}{
		{
			test:    "no route table associations (test for nil values)",
			dirName: "route_table_assoc_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllRouteTables").Once().Return([]*ec2.RouteTable{
					{
						RouteTableId: awssdk.String("assoc_with_nil"),
						Associations: []*ec2.RouteTableAssociation{
							{
								AssociationState:        nil,
								GatewayId:               nil,
								Main:                    nil,
								RouteTableAssociationId: nil,
								RouteTableId:            nil,
								SubnetId:                nil,
							},
						},
					},
					{
						RouteTableId: awssdk.String("nil_assoc"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "route_table_association (mixed subnet and gateway associations)",
			dirName: "route_table_assoc",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllRouteTables").Once().Return([]*ec2.RouteTable{
					{
						RouteTableId: aws.String("rtb-05aa6c5673311a17b"), // route
						Associations: []*ec2.RouteTableAssociation{
							{ // Should be ignored
								AssociationState: &ec2.RouteTableAssociationState{
									State: awssdk.String("disassociated"),
								},
								GatewayId: awssdk.String("dummy-id"),
							},
							{ // Should be ignored
								SubnetId:  nil,
								GatewayId: nil,
							},
							{ // assoc_route_subnet1
								AssociationState: &ec2.RouteTableAssociationState{
									State: awssdk.String("associated"),
								},
								Main:                    awssdk.Bool(false),
								RouteTableAssociationId: awssdk.String("rtbassoc-0809598f92dbec03b"),
								RouteTableId:            awssdk.String("rtb-05aa6c5673311a17b"),
								SubnetId:                awssdk.String("subnet-05185af647b2eeda3"),
							},
							{ // assoc_route_subnet
								AssociationState: &ec2.RouteTableAssociationState{
									State: awssdk.String("associated"),
								},
								Main:                    awssdk.Bool(false),
								RouteTableAssociationId: awssdk.String("rtbassoc-01957791b2cfe6ea4"),
								RouteTableId:            awssdk.String("rtb-05aa6c5673311a17b"),
								SubnetId:                awssdk.String("subnet-0e93dbfa2e5dd8282"),
							},
							{ // assoc_route_subnet2
								AssociationState: &ec2.RouteTableAssociationState{
									State: awssdk.String("associated"),
								},
								GatewayId:               nil,
								Main:                    awssdk.Bool(false),
								RouteTableAssociationId: awssdk.String("rtbassoc-0b4f97ea57490e213"),
								RouteTableId:            awssdk.String("rtb-05aa6c5673311a17b"),
								SubnetId:                awssdk.String("subnet-0fd966efd884d0362"),
							},
						},
					},
					{
						RouteTableId: aws.String("rtb-09df7cc9d16de9f8f"), // route2
						Associations: []*ec2.RouteTableAssociation{
							{ // assoc_route2_gateway
								AssociationState: &ec2.RouteTableAssociationState{
									State: awssdk.String("associated"),
								},
								RouteTableAssociationId: awssdk.String("rtbassoc-0a79ccacfceb4944b"),
								RouteTableId:            awssdk.String("rtb-09df7cc9d16de9f8f"),
								GatewayId:               awssdk.String("igw-0238f6e09185ac954"),
							},
						},
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "Cannot list route table",
			dirName: "route_table_assoc_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllRouteTables").Once().Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsRouteTableAssociationResourceType, resourceaws.AwsRouteTableResourceType),
		},
	}
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := InitTestAwsProvider(providerLibrary)
			if err != nil {
				t.Fatal(err)
			}
			supplierLibrary.AddSupplier(NewRouteTableAssociationSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeEC2 := repository.MockEC2Repository{}
			c.mocks(&fakeEC2)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			routeTableAssociationDeserializer := awsdeserializer.NewRouteTableAssociationDeserializer()
			s := &RouteTableAssociationSupplier{
				provider,
				routeTableAssociationDeserializer,
				&fakeEC2,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			deserializers := []deserializer.CTYDeserializer{routeTableAssociationDeserializer}
			test.CtyTestDiffMixed(got, c.dirName, provider, deserializers, shouldUpdate, tt)
		})
	}
}

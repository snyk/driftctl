package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	awssdk "github.com/aws/aws-sdk-go/aws"
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

func TestRouteTableSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockEC2Repository)
		err     error
	}{
		{
			test:    "no route table",
			dirName: "route_table_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllRouteTables").Once().Return([]*ec2.RouteTable{}, nil)
			},
			err: nil,
		},
		{
			test:    "mixed default_route_table and route_table",
			dirName: "route_table",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllRouteTables").Once().Return([]*ec2.RouteTable{
					{
						RouteTableId: aws.String("rtb-08b7b71af15e183ce"), // table1
					},
					{
						RouteTableId: aws.String("rtb-0002ac731f6fdea55"), // table2
					},
					{
						RouteTableId: aws.String("rtb-0eabf071c709c0976"), // default_table
						VpcId:        awssdk.String("vpc-0b4a6b3536da20ecd"),
						Associations: []*ec2.RouteTableAssociation{
							{
								Main: awssdk.Bool(true),
							},
						},
					},
					{
						RouteTableId: aws.String("rtb-0c55d55593f33fbac"), // table3
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list route table",
			dirName: "route_table_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllRouteTables").Once().Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsRouteTableResourceType),
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
			supplierLibrary.AddSupplier(NewRouteTableSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeEC2 := repository.MockEC2Repository{}
			c.mocks(&fakeEC2)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			routeTableDeserializer := awsdeserializer.NewRouteTableDeserializer()
			defaultRouteTableDeserializer := awsdeserializer.NewDefaultRouteTableDeserializer()
			s := &RouteTableSupplier{
				provider,
				defaultRouteTableDeserializer,
				routeTableDeserializer,
				&fakeEC2,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			deserializers := []deserializer.CTYDeserializer{routeTableDeserializer, defaultRouteTableDeserializer}
			test.CtyTestDiffMixed(got, c.dirName, provider, deserializers, shouldUpdate, tt)
		})
	}
}

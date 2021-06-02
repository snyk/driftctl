package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	awstest "github.com/cloudskiff/driftctl/test/aws"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNatGatewaySupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *awstest.MockFakeEC2)
		err     error
	}{
		{
			test:    "no gateway",
			dirName: "nat_gateway_empty",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeNatGatewaysPages",
					&ec2.DescribeNatGatewaysInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeNatGatewaysOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeNatGatewaysOutput{}, true)
						return true
					})).Return(nil)
			},
			err: nil,
		},
		{
			test:    "single aws_nat_gateway",
			dirName: "nat_gateway",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeNatGatewaysPages",
					&ec2.DescribeNatGatewaysInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeNatGatewaysOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeNatGatewaysOutput{
							NatGateways: []*ec2.NatGateway{
								{
									NatGatewayId: aws.String("nat-0a5408508b19ef490"), // nat1
								},
							},
						}, true)
						return true
					})).Return(nil)
			},
			err: nil,
		},
		{
			test:    "cannot list gateway",
			dirName: "nat_gateway_empty",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeNatGatewaysPages",
					&ec2.DescribeNatGatewaysInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeNatGatewaysOutput, lastPage bool) bool) bool {
						return true
					})).Return(awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsNatGatewayResourceType),
		},
	}
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
		resourceaws.InitResourcesMetadata(repo)
		factory := terraform.NewTerraformResourceFactory(repo)

		deserializer := resource.NewDeserializer(factory)
		if shouldUpdate {
			provider, err := InitTestAwsProvider(providerLibrary)
			if err != nil {
				t.Fatal(err)
			}
			supplierLibrary.AddSupplier(NewNatGatewaySupplier(provider, deserializer))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeEC2 := awstest.MockFakeEC2{}
			c.mocks(&fakeEC2)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			s := &NatGatewaySupplier{
				provider,
				deserializer,
				&fakeEC2,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

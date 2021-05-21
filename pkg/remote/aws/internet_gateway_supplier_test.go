package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	awstest "github.com/cloudskiff/driftctl/test/aws"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	awssdk "github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/resource"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInternetGatewaySupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *awstest.MockFakeEC2)
		err     error
	}{
		{
			test:    "no internet gateways",
			dirName: "internet_gateway_empty",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeInternetGatewaysPages",
					&ec2.DescribeInternetGatewaysInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeInternetGatewaysOutput{}, true)
						return true
					})).Return(nil)
			},
			err: nil,
		},
		{
			test:    "multiple internet gateways",
			dirName: "internet_gateway_multiple",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeInternetGatewaysPages",
					&ec2.DescribeInternetGatewaysInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeInternetGatewaysOutput{
							InternetGateways: []*ec2.InternetGateway{
								{
									InternetGatewayId: awssdk.String("igw-0184eb41aadc62d1c"),
								},
								{
									InternetGatewayId: awssdk.String("igw-047b487f5c60fca99"),
								},
							},
						}, true)
						return true
					})).Return(nil)
			},
			err: nil,
		},
		{
			test:    "cannot list internet gateways",
			dirName: "internet_gateway_empty",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeInternetGatewaysPages",
					&ec2.DescribeInternetGatewaysInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool) bool {
						return true
					})).Return(awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsInternetGatewayResourceType),
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
			supplierLibrary.AddSupplier(NewInternetGatewaySupplier(provider, deserializer))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeEC2 := awstest.MockFakeEC2{}
			c.mocks(&fakeEC2)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			s := &InternetGatewaySupplier{
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

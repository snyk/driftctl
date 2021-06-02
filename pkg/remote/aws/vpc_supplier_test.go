package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	awstest "github.com/cloudskiff/driftctl/test/aws"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	"github.com/aws/aws-sdk-go/aws/awserr"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
)

func TestVPCSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *awstest.MockFakeEC2)
		err     error
	}{
		{
			test:    "no VPC",
			dirName: "vpc_empty",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeVpcsPages",
					&ec2.DescribeVpcsInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeVpcsOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeVpcsOutput{}, true)
						return true
					})).Return(nil)
			},
			err: nil,
		},
		{
			test:    "mixed default VPC and VPC",
			dirName: "vpc",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeVpcsPages",
					&ec2.DescribeVpcsInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeVpcsOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeVpcsOutput{
							Vpcs: []*ec2.Vpc{
								{
									VpcId:     aws.String("vpc-a8c5d4c1"),
									IsDefault: aws.Bool(true),
								},
								{
									VpcId: aws.String("vpc-0768e1fd0029e3fc3"),
								},
								{
									VpcId:     aws.String("vpc-020b072316a95b97f"),
									IsDefault: aws.Bool(false),
								},
							},
						}, false)
						callback(&ec2.DescribeVpcsOutput{
							Vpcs: []*ec2.Vpc{
								{
									VpcId:     aws.String("vpc-02c50896b59598761"),
									IsDefault: aws.Bool(false),
								},
							},
						}, true)
						return true
					})).Return(nil)
			},
			err: nil,
		},
		{
			test:    "cannot list VPC",
			dirName: "vpc_empty",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeVpcsPages",
					&ec2.DescribeVpcsInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeVpcsOutput, lastPage bool) bool) bool {
						return true
					})).Return(awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsVpcResourceType),
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
			supplierLibrary.AddSupplier(NewVPCSupplier(provider, deserializer))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeEC2 := awstest.MockFakeEC2{}
			c.mocks(&fakeEC2)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			s := &VPCSupplier{
				provider,
				deserializer,
				&fakeEC2,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	awstest "github.com/cloudskiff/driftctl/test/aws"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

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

func TestSubnetSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *awstest.MockFakeEC2)
		err     error
	}{
		{
			test:    "no Subnet",
			dirName: "subnet_empty",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeSubnetsPages",
					&ec2.DescribeSubnetsInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeSubnetsOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeSubnetsOutput{}, true)
						return true
					})).Return(nil)
			},
			err: nil,
		},
		{
			test:    "mixed default Subnet and Subnet",
			dirName: "subnet",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeSubnetsPages",
					&ec2.DescribeSubnetsInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeSubnetsOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeSubnetsOutput{
							Subnets: []*ec2.Subnet{
								{
									SubnetId:     aws.String("subnet-44fe0c65"), // us-east-1a
									DefaultForAz: aws.Bool(true),
								},
								{
									SubnetId:     aws.String("subnet-65e16628"), // us-east-1b
									DefaultForAz: aws.Bool(true),
								},
								{
									SubnetId:     aws.String("subnet-afa656f0"), // us-east-1c
									DefaultForAz: aws.Bool(true),
								},
								{
									SubnetId:     aws.String("subnet-05810d3f933925f6d"), // subnet1
									DefaultForAz: aws.Bool(false),
								},
							},
						}, false)
						callback(&ec2.DescribeSubnetsOutput{
							Subnets: []*ec2.Subnet{
								{
									SubnetId:     aws.String("subnet-0b13f1e0eacf67424"), // subnet2
									DefaultForAz: aws.Bool(false),
								},
								{
									SubnetId:     aws.String("subnet-0c9b78001fe186e22"), // subnet3
									DefaultForAz: aws.Bool(false),
								},
							},
						}, true)
						return true
					})).Return(nil)
			},
			err: nil,
		},
		{
			test:    "cannot list Subnet",
			dirName: "subnet_empty",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeSubnetsPages",
					&ec2.DescribeSubnetsInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeSubnetsOutput, lastPage bool) bool) bool {
						return true
					})).Return(awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsSubnetResourceType),
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
			supplierLibrary.AddSupplier(NewSubnetSupplier(provider, deserializer))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeEC2 := awstest.MockFakeEC2{}
			c.mocks(&fakeEC2)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			s := &SubnetSupplier{
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

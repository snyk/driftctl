package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/aws/aws-sdk-go/aws/awserr"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/mocks"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
)

func TestVPCSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *mocks.FakeEC2)
		err     error
	}{
		{
			test:    "no VPC",
			dirName: "vpc_empty",
			mocks: func(client *mocks.FakeEC2) {
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
			mocks: func(client *mocks.FakeEC2) {
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
			mocks: func(client *mocks.FakeEC2) {
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

		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			providerLibrary.AddProvider(terraform.AWS, provider)
			supplierLibrary.AddSupplier(NewVPCSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeEC2 := mocks.FakeEC2{}
			c.mocks(&fakeEC2)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			VPCDeserializer := awsdeserializer.NewVPCDeserializer()
			defaultVPCDeserializer := awsdeserializer.NewDefaultVPCDeserializer()
			s := &VPCSupplier{
				provider,
				defaultVPCDeserializer,
				VPCDeserializer,
				&fakeEC2,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			deserializers := []deserializer.CTYDeserializer{VPCDeserializer, defaultVPCDeserializer}
			test.CtyTestDiffMixed(got, c.dirName, provider, deserializers, shouldUpdate, tt)
		})
	}
}

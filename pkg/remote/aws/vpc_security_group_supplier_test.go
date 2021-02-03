package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/aws/aws-sdk-go/aws/awserr"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVPCSecurityGroupSupplier_Resources(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(client *mocks.FakeEC2)
		err     error
	}{
		{
			test:    "no security groups",
			dirName: "vpc_security_group_empty",
			mocks: func(client *mocks.FakeEC2) {
				client.On("DescribeSecurityGroupsPages",
					&ec2.DescribeSecurityGroupsInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeSecurityGroupsOutput{}, true)
						return true
					})).Return(nil)
			},
			err: nil,
		},
		{
			test:    "with security groups",
			dirName: "vpc_security_group_multiple",
			mocks: func(client *mocks.FakeEC2) {
				client.On("DescribeSecurityGroupsPages",
					&ec2.DescribeSecurityGroupsInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []*ec2.SecurityGroup{
								{
									GroupId:   aws.String("sg-0254c038e32f25530"),
									GroupName: aws.String("foo"),
								},
								{
									GroupId:   aws.String("sg-9e0204ff"),
									GroupName: aws.String("default"),
								},
							},
						}, true)
						return true
					})).Return(nil)
			},
			err: nil,
		},
		{
			test:    "cannot list security groups",
			dirName: "vpc_security_group_empty",
			mocks: func(client *mocks.FakeEC2) {
				client.On("DescribeSecurityGroupsPages",
					&ec2.DescribeSecurityGroupsInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool) bool {
						return true
					})).Return(awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsSecurityGroupResourceType),
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			providerLibrary.AddProvider(terraform.AWS, provider)
			supplierLibrary.AddSupplier(NewVPCSecurityGroupSupplier(provider))
		}

		t.Run(tt.test, func(t *testing.T) {
			fakeEC2 := mocks.FakeEC2{}
			tt.mocks(&fakeEC2)
			provider := mocks2.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			securityGroupDeserializer := awsdeserializer.NewVPCSecurityGroupDeserializer()
			defaultSecurityGroupDeserializer := awsdeserializer.NewDefaultSecurityGroupDeserializer()
			s := &VPCSecurityGroupSupplier{
				provider,
				defaultSecurityGroupDeserializer,
				securityGroupDeserializer,
				&fakeEC2,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(t, tt.err, err)

			mock.AssertExpectationsForObjects(t)
			deserializers := []deserializer.CTYDeserializer{securityGroupDeserializer, defaultSecurityGroupDeserializer}
			test.CtyTestDiffMixed(got, tt.dirName, provider, deserializers, shouldUpdate, t)
		})
	}
}

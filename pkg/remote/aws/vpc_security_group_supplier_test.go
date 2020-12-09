package aws

import (
	"context"
	"testing"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestVPCSecurityGroupSupplier_Resources(t *testing.T) {
	tests := []struct {
		test                string
		dirName             string
		securityGroupsPages mocks.DescribeSecurityGroupsPagesOutput
		err                 error
	}{
		{
			test:    "no security groups",
			dirName: "vpc_security_group_empty",
			securityGroupsPages: mocks.DescribeSecurityGroupsPagesOutput{
				{
					true,
					&ec2.DescribeSecurityGroupsOutput{},
				},
			},
			err: nil,
		},
		{
			test:    "with security groups",
			dirName: "vpc_security_group_multiple",
			securityGroupsPages: mocks.DescribeSecurityGroupsPagesOutput{
				{
					false,
					&ec2.DescribeSecurityGroupsOutput{
						SecurityGroups: []*ec2.SecurityGroup{
							{
								GroupId: aws.String("sg-0254c038e32f25530"),
							},
						},
					},
				},
				{
					true,
					&ec2.DescribeSecurityGroupsOutput{
						SecurityGroups: []*ec2.SecurityGroup{
							{
								GroupId: aws.String("sg-9e0204ff"),
							},
						},
					},
				},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update
		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			terraform.AddProvider(terraform.AWS, provider)
			resource.AddSupplier(NewVPCSecurityGroupSupplier(provider.Runner(), ec2.New(provider.session)))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, terraform.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewVPCSecurityGroupDeserializer()
			s := &VPCSecurityGroupSupplier{
				provider,
				deserializer,
				mocks.NewMockAWSVPCSecurityGroupClient(tt.securityGroupsPages),
				terraform.NewParallelResourceReader(pkg.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			if tt.err != err {
				t.Errorf("Expected error %+v got %+v", tt.err, err)
			}

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}

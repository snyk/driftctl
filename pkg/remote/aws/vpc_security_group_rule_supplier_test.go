package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/aws/aws-sdk-go/aws/awserr"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

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

func TestVPCSecurityGroupRuleSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockEC2Repository)
		err     error
	}{
		{
			test:    "no security group rules",
			dirName: "vpc_security_group_rule_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllSecurityGroups").Once().Return([]*ec2.SecurityGroup{
					{
						GroupId:             aws.String("sg-0254c038e32f25530"),
						IpPermissions:       []*ec2.IpPermission{},
						IpPermissionsEgress: []*ec2.IpPermission{},
					},
				}, nil, nil)
			},
			err: nil,
		},
		{
			test:    "with security group rules",
			dirName: "vpc_security_group_rule_multiple",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllSecurityGroups").Once().Return([]*ec2.SecurityGroup{
					{
						GroupId: aws.String("sg-0254c038e32f25530"),
						IpPermissions: []*ec2.IpPermission{
							{
								FromPort:   aws.Int64(0),
								ToPort:     aws.Int64(65535),
								IpProtocol: aws.String("tcp"),
								UserIdGroupPairs: []*ec2.UserIdGroupPair{
									{
										GroupId: aws.String("sg-0254c038e32f25530"),
									},
									{
										GroupId: aws.String("sg-9e0204ff"),
									},
								},
							},
							{
								IpProtocol: aws.String("-1"),
								IpRanges: []*ec2.IpRange{
									{
										CidrIp: aws.String("1.2.0.0/16"),
									},
									{
										CidrIp: aws.String("5.6.7.0/24"),
									},
								},
								Ipv6Ranges: []*ec2.Ipv6Range{
									{
										CidrIpv6: aws.String("::/0"),
									},
								},
							},
						},
						IpPermissionsEgress: []*ec2.IpPermission{
							{
								IpProtocol: aws.String("-1"),
								IpRanges: []*ec2.IpRange{
									{
										CidrIp: aws.String("0.0.0.0/0"),
									},
								},
								Ipv6Ranges: []*ec2.Ipv6Range{
									{
										CidrIpv6: aws.String("::/0"),
									},
								},
							},
						},
					},
					{
						GroupId: aws.String("sg-0cc8b3c3c2851705a"),
						IpPermissions: []*ec2.IpPermission{
							{
								FromPort:   aws.Int64(443),
								ToPort:     aws.Int64(443),
								IpProtocol: aws.String("tcp"),
								IpRanges: []*ec2.IpRange{
									{
										CidrIp: aws.String("0.0.0.0/0"),
									},
								},
							},
						},
						IpPermissionsEgress: []*ec2.IpPermission{
							{
								IpProtocol: aws.String("-1"),
								IpRanges: []*ec2.IpRange{
									{
										CidrIp: aws.String("0.0.0.0/0"),
									},
								},
								Ipv6Ranges: []*ec2.Ipv6Range{
									{
										CidrIpv6: aws.String("::/0"),
									},
								},
							},
							{
								IpProtocol: aws.String("5"),
								IpRanges: []*ec2.IpRange{
									{
										CidrIp: aws.String("0.0.0.0/0"),
									},
								},
							},
						},
					},
				}, nil, nil)
			},
			err: nil,
		},
		{
			test:    "should ignore default security group default rules",
			dirName: "vpc_security_group_default_rules",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllSecurityGroups").Once().Return([]*ec2.SecurityGroup{
					{
						GroupId:   aws.String("sg-a74815c8"),
						GroupName: aws.String("default"),
						IpPermissions: []*ec2.IpPermission{
							{
								IpProtocol: aws.String("-1"),
								UserIdGroupPairs: []*ec2.UserIdGroupPair{
									{
										GroupId: aws.String("sg-a74815c8"),
									},
								},
							},
							{
								IpProtocol: aws.String("-1"),
								IpRanges: []*ec2.IpRange{
									{
										CidrIp: aws.String("1.2.0.0/16"),
									},
								},
							},
						},
						IpPermissionsEgress: []*ec2.IpPermission{
							{
								IpProtocol: aws.String("-1"),
								IpRanges: []*ec2.IpRange{
									{
										CidrIp: aws.String("0.0.0.0/0"),
									},
								},
							},
							{
								IpProtocol: aws.String("-1"),
								IpRanges: []*ec2.IpRange{
									{
										CidrIp: aws.String("1.2.3.4/32"),
									},
								},
							},
						},
					},
				}, nil, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list security group rules",
			dirName: "vpc_security_group_rule_empty",
			mocks: func(client *repository.MockEC2Repository) {
				client.On("ListAllSecurityGroups").Once().Return(nil, nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsSecurityGroupRuleResourceType),
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
			supplierLibrary.AddSupplier(NewVPCSecurityGroupRuleSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeEC2 := repository.MockEC2Repository{}
			c.mocks(&fakeEC2)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewVPCSecurityGroupRuleDeserializer()
			s := &VPCSecurityGroupRuleSupplier{
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

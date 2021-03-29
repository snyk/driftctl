package middlewares

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

func TestVPCSecurityGroupRuleSanitizer(t *testing.T) {

	factory := &terraform.MockResourceFactory{}
	factory.On("CreateResource", mock.Anything, "aws_security_group_rule").Times(8).Return(nil, nil)

	middleware := NewVPCSecurityGroupRuleSanitizer(factory)
	var remoteResources []resource.Resource
	stateResources := []resource.Resource{
		&aws.AwsSecurityGroup{
			Id:   "sg-test",
			Name: awssdk.String("test"),
		},
		&aws.AwsSecurityGroupRule{
			Id:                    "sgrule-3970541193",
			Type:                  awssdk.String("ingress"),
			SecurityGroupId:       awssdk.String("sg-0254c038e32f25530"),
			Protocol:              awssdk.String("tcp"),
			FromPort:              awssdk.Int(0),
			ToPort:                awssdk.Int(65535),
			Self:                  awssdk.Bool(true),
			SourceSecurityGroupId: awssdk.String("sg-0254c038e32f25530"),
		},
		&aws.AwsSecurityGroupRule{
			Id:              "sgrule-845917806",
			Type:            awssdk.String("egress"),
			SecurityGroupId: awssdk.String("sg-0cc8b3c3c2851705a"),
			Protocol:        awssdk.String("-1"),
			FromPort:        awssdk.Int(0),
			ToPort:          awssdk.Int(0),
			CidrBlocks:      &[]string{"0.0.0.0/0"},
			Ipv6CidrBlocks:  &[]string{"::/0"},
		},
		&aws.AwsSecurityGroupRule{
			Id:              "sgrule-294318973",
			Type:            awssdk.String("ingress"),
			SecurityGroupId: awssdk.String("sg-0254c038e32f25530"),
			Protocol:        awssdk.String("-1"),
			FromPort:        awssdk.Int(0),
			ToPort:          awssdk.Int(0),
			CidrBlocks:      &[]string{"1.2.0.0/16", "5.6.7.0/24"},
		},
		&aws.AwsSecurityGroupRule{
			Id:              "sgrule-2471889226",
			Type:            awssdk.String("ingress"),
			SecurityGroupId: awssdk.String("sg-0254c038e32f25530"),
			Protocol:        awssdk.String("tcp"),
			FromPort:        awssdk.Int(0),
			ToPort:          awssdk.Int(0),
			PrefixListIds:   &[]string{"pl-abb451c2"},
		},
		&aws.AwsSecurityGroupRule{
			Id:                    "sgrule-3587309474",
			Type:                  awssdk.String("ingress"),
			SecurityGroupId:       awssdk.String("sg-0254c038e32f25530"),
			Protocol:              awssdk.String("tcp"),
			FromPort:              awssdk.Int(0),
			ToPort:                awssdk.Int(65535),
			SourceSecurityGroupId: awssdk.String("sg-9e0204ff"),
		},
	}
	err := middleware.Execute(&remoteResources, &stateResources)
	if err != nil {
		t.Error(err)
	}
	if len(stateResources) != 8 {
		t.Error("Some security group rules were not split")
	}
}

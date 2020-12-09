package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsSecurityGroupRule_String(t *testing.T) {
	tests := []struct {
		name string
		rule AwsSecurityGroupRule
		want string
	}{
		{
			name: "Test stringer on ingress rule",
			rule: AwsSecurityGroupRule{
				CidrBlocks:      &[]string{"0.0.0.0/0", "1.2.3.4/32"},
				FromPort:        aws.Int(22),
				Protocol:        aws.String("tcp"),
				SecurityGroupId: aws.String("sg-12345"),
				ToPort:          aws.Int(22),
				Type:            aws.String("ingress"),
			},
			want: "Type: ingress, SecurityGroup: sg-12345, Protocol: tcp, Ports: 22, Source: 0.0.0.0/0, 1.2.3.4/32",
		},
		{
			name: "Test stringer on egress rule",
			rule: AwsSecurityGroupRule{
				CidrBlocks:      &[]string{"0.0.0.0/0", "1.2.3.4/32"},
				FromPort:        aws.Int(22),
				Protocol:        aws.String("tcp"),
				SecurityGroupId: aws.String("sg-12345"),
				ToPort:          aws.Int(22),
				Type:            aws.String("egress"),
			},
			want: "Type: egress, SecurityGroup: sg-12345, Protocol: tcp, Ports: 22, Destination: 0.0.0.0/0, 1.2.3.4/32",
		},
		{
			name: "Test protocol display 'All' and empty cidr",
			rule: AwsSecurityGroupRule{
				Protocol: aws.String("-1"),
			},
			want: "Protocol: All",
		},
		{
			name: "Test port range",
			rule: AwsSecurityGroupRule{
				FromPort: aws.Int(22),
				ToPort:   aws.Int(25),
			},
			want: "Ports: 22-25",
		},
		{
			name: "Test port range show all",
			rule: AwsSecurityGroupRule{
				FromPort: aws.Int(0),
				ToPort:   aws.Int(0),
			},
			want: "Ports: All",
		},
		{
			name: "Test empty cidr",
			rule: AwsSecurityGroupRule{
				CidrBlocks: &[]string{},
			},
			want: "",
		},
		{
			name: "Test empty cidrv6",
			rule: AwsSecurityGroupRule{
				Ipv6CidrBlocks: &[]string{},
			},
			want: "",
		},
		{
			name: "Test cidr v6",
			rule: AwsSecurityGroupRule{
				Ipv6CidrBlocks: &[]string{"::/0"},
				Type:           aws.String("ingress"),
			},
			want: "Type: ingress, Source: ::/0",
		},
		{
			name: "Test empty prefix list",
			rule: AwsSecurityGroupRule{
				PrefixListIds: &[]string{},
			},
			want: "",
		},
		{
			name: "Test prefix list",
			rule: AwsSecurityGroupRule{
				PrefixListIds: &[]string{"pl-12345"},
				Type:          aws.String("egress"),
			},
			want: "Type: egress, Destination: pl-12345",
		},
		{
			name: "Test source security group id",
			rule: AwsSecurityGroupRule{
				SourceSecurityGroupId: aws.String("sg-1234"),
				Type:                  aws.String("egress"),
			},
			want: "Type: egress, Destination: sg-1234",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rule.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

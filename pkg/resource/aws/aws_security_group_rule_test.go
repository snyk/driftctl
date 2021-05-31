package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAwsSecurityGroupRule_Attrs(t *testing.T) {
	tests := []struct {
		name string
		rule AwsSecurityGroupRule
		want map[string]string
	}{
		{
			name: "Test attrs on ingress rule",
			rule: AwsSecurityGroupRule{
				CidrBlocks:      &[]string{"0.0.0.0/0", "1.2.3.4/32"},
				FromPort:        aws.Int(22),
				Protocol:        aws.String("tcp"),
				SecurityGroupId: aws.String("sg-12345"),
				ToPort:          aws.Int(22),
				Type:            aws.String("ingress"),
			},
			want: map[string]string{
				"SecurityGroup": "sg-12345",
				"Protocol":      "tcp",
				"Ports":         "22",
				"Source":        "0.0.0.0/0, 1.2.3.4/32",
				"Type":          "ingress",
			},
		},
		{
			name: "Test attrs on egress rule",
			rule: AwsSecurityGroupRule{
				CidrBlocks:      &[]string{"0.0.0.0/0", "1.2.3.4/32"},
				FromPort:        aws.Int(22),
				Protocol:        aws.String("tcp"),
				SecurityGroupId: aws.String("sg-12345"),
				ToPort:          aws.Int(22),
				Type:            aws.String("egress"),
			},
			want: map[string]string{
				"Type":          "egress",
				"SecurityGroup": "sg-12345",
				"Protocol":      "tcp",
				"Ports":         "22",
				"Destination":   "0.0.0.0/0, 1.2.3.4/32",
			},
		},
		{
			name: "Test protocol display 'All' and empty cidr",
			rule: AwsSecurityGroupRule{
				Protocol:        aws.String("-1"),
				SecurityGroupId: aws.String("sg-12345"),
				Type:            aws.String("ingress"),
			},
			want: map[string]string{
				"Type":          "ingress",
				"SecurityGroup": "sg-12345",
				"Protocol":      "All",
			},
		},
		{
			name: "Test port range",
			rule: AwsSecurityGroupRule{
				FromPort:        aws.Int(22),
				ToPort:          aws.Int(25),
				SecurityGroupId: aws.String("sg-12345"),
				Protocol:        aws.String("-1"),
				Type:            aws.String("ingress"),
			},
			want: map[string]string{
				"Type":          "ingress",
				"SecurityGroup": "sg-12345",
				"Protocol":      "All",
				"Ports":         "22-25",
			},
		},
		{
			name: "Test port range show all",
			rule: AwsSecurityGroupRule{
				FromPort:        aws.Int(0),
				ToPort:          aws.Int(0),
				SecurityGroupId: aws.String("sg-12345"),
				Protocol:        aws.String("-1"),
				Type:            aws.String("ingress"),
			},
			want: map[string]string{
				"Type":          "ingress",
				"SecurityGroup": "sg-12345",
				"Protocol":      "All",
				"Ports":         "All",
			},
		},
		{
			name: "Test empty cidr",
			rule: AwsSecurityGroupRule{
				CidrBlocks:      &[]string{},
				SecurityGroupId: aws.String("sg-12345"),
				Protocol:        aws.String("-1"),
				Type:            aws.String("ingress"),
			},
			want: map[string]string{
				"Type":          "ingress",
				"SecurityGroup": "sg-12345",
				"Protocol":      "All",
			},
		},
		{
			name: "Test empty cidrv6",
			rule: AwsSecurityGroupRule{
				Ipv6CidrBlocks:  &[]string{},
				SecurityGroupId: aws.String("sg-12345"),
				Protocol:        aws.String("-1"),
				Type:            aws.String("ingress"),
			},
			want: map[string]string{
				"Type":          "ingress",
				"SecurityGroup": "sg-12345",
				"Protocol":      "All",
			},
		},
		{
			name: "Test cidr v6",
			rule: AwsSecurityGroupRule{
				Ipv6CidrBlocks:  &[]string{"::/0"},
				Type:            aws.String("ingress"),
				SecurityGroupId: aws.String("sg-12345"),
				Protocol:        aws.String("-1"),
			},
			want: map[string]string{
				"Type":          "ingress",
				"SecurityGroup": "sg-12345",
				"Protocol":      "All",
				"Source":        "::/0",
			},
		},
		{
			name: "Test empty prefix list",
			rule: AwsSecurityGroupRule{
				PrefixListIds:   &[]string{},
				SecurityGroupId: aws.String("sg-12345"),
				Protocol:        aws.String("-1"),
				Type:            aws.String("ingress"),
			},
			want: map[string]string{
				"Type":          "ingress",
				"SecurityGroup": "sg-12345",
				"Protocol":      "All",
			},
		},
		{
			name: "Test prefix list",
			rule: AwsSecurityGroupRule{
				PrefixListIds:   &[]string{"pl-12345"},
				Type:            aws.String("egress"),
				SecurityGroupId: aws.String("sg-12345"),
				Protocol:        aws.String("-1"),
			},
			want: map[string]string{
				"Type":          "egress",
				"SecurityGroup": "sg-12345",
				"Protocol":      "All",
				"Destination":   "pl-12345",
			},
		},
		{
			name: "Test source security group id",
			rule: AwsSecurityGroupRule{
				SourceSecurityGroupId: aws.String("sg-1234"),
				Type:                  aws.String("egress"),
				SecurityGroupId:       aws.String("sg-12345"),
				Protocol:              aws.String("-1"),
			},
			want: map[string]string{
				"Type":          "egress",
				"SecurityGroup": "sg-12345",
				"Protocol":      "All",
				"Destination":   "sg-12345",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rule.Attributes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Attributes() = %v, want %v", got, tt.want)
			}
		})
	}
}

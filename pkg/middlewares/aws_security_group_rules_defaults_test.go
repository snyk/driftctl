package middlewares

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

func TestAwsSecurityGroupRuleDefaults_Execute(t *testing.T) {
	defaultSecurityGroupName := "default"
	defaultSecurityGroupId := "sg-test1"
	defaultSecurityGroupRuleProtocol := "All"
	defaultSecurityGroupRuleType := "ingress"
	defaultSecurityGroupRuleDescription := "test desc"

	dummySecurityGroupName := "sg-test2"
	dummySecurityGroupId := "sg-test2"

	tests := []struct {
		name               string
		remoteResources    []resource.Resource
		resourcesFromState []resource.Resource
		expected           diff.Changelog
	}{
		{
			"default security group when they're not managed by IaC",
			[]resource.Resource{
				&aws.AwsSecurityGroup{
					Id:   defaultSecurityGroupId,
					Name: &defaultSecurityGroupName,
				},
				&aws.AwsSecurityGroupRule{
					Id:              "test-1",
					SecurityGroupId: &defaultSecurityGroupId,
					Type:            &defaultSecurityGroupRuleType,
					Protocol:        &defaultSecurityGroupRuleProtocol,
				},
				&aws.AwsRoute{
					Id:           "dummy-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("local"),
				},
			},
			[]resource.Resource{
				&aws.AwsRoute{
					Id:           "dummy-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("local"),
				},
			},
			diff.Changelog{
				{
					Type: "delete",
					Path: []string{"0"},
					From: &aws.AwsSecurityGroup{
						Id:   defaultSecurityGroupId,
						Name: &defaultSecurityGroupName,
					},
					To: nil,
				},
			},
		},
		{
			"default security group when they're managed by IaC",
			[]resource.Resource{
				&aws.AwsSecurityGroup{
					Id:   defaultSecurityGroupId,
					Name: &defaultSecurityGroupName,
				},
				&aws.AwsSecurityGroupRule{
					Id:              "test-1",
					SecurityGroupId: &defaultSecurityGroupId,
					Type:            &defaultSecurityGroupRuleType,
					Protocol:        &defaultSecurityGroupRuleProtocol,
					Description:     nil,
				},
				&aws.AwsSecurityGroup{
					Id:   dummySecurityGroupId,
					Name: &dummySecurityGroupName,
				},
			},
			[]resource.Resource{
				&aws.AwsSecurityGroup{
					Id:   defaultSecurityGroupId,
					Name: &defaultSecurityGroupName,
				},
				&aws.AwsSecurityGroupRule{
					Id:              "test-1",
					SecurityGroupId: &defaultSecurityGroupId,
					Type:            &defaultSecurityGroupRuleType,
					Protocol:        &defaultSecurityGroupRuleProtocol,
					Description:     &defaultSecurityGroupRuleDescription,
				},
			},
			diff.Changelog{
				{
					Type: "delete",
					Path: []string{"2"},
					From: &aws.AwsSecurityGroup{
						Id:   dummySecurityGroupId,
						Name: &dummySecurityGroupName,
					},
					To: nil,
				},
				{
					Type: "update",
					Path: []string{"1", "Description"},
					From: nil,
					To:   &defaultSecurityGroupRuleDescription,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsSecurityGroupRuleDefaults()
			err := m.Execute(&tt.remoteResources, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}

			changelog, err := diff.Diff(tt.remoteResources, tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}

			diffs, err := diff.Diff(tt.expected, changelog)
			if err != nil {
				t.Fatal(err)
			}

			for _, change := range diffs {
				t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
			}
		})
	}
}

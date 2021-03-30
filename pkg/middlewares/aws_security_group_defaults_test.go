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

func TestAwsSecurityGroupDefaults_Execute(t *testing.T) {
	defaultSecurityGroupName := "default"
	dummySecurityGroupName := "test-group"
	dummySecurityGroupDescription := "test-desc"

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
					Id:   "sg-51530134",
					Name: &defaultSecurityGroupName,
				},
				&aws.AwsSecurityGroup{
					Id:   "test",
					Name: &dummySecurityGroupName,
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
						Id:   "test",
						Name: &dummySecurityGroupName,
					},
					To: nil,
				},
			},
		},
		{
			"default security group when they're managed by IaC",
			[]resource.Resource{
				&aws.AwsSecurityGroup{
					Id:   "sg-51530134",
					Name: &defaultSecurityGroupName,
				},
				&aws.AwsSecurityGroup{
					Id:   "test",
					Name: &dummySecurityGroupName,
				},
			},
			[]resource.Resource{
				&aws.AwsSecurityGroup{
					Id:          "sg-51530134",
					Name:        &defaultSecurityGroupName,
					Description: &dummySecurityGroupDescription,
				},
			},
			diff.Changelog{
				{
					Type: "update",
					Path: []string{"0", "Description"},
					From: nil,
					To:   &dummySecurityGroupDescription,
				},
				{
					Type: "delete",
					Path: []string{"1"},
					From: &aws.AwsSecurityGroup{
						Id:   "test",
						Name: &dummySecurityGroupName,
					},
					To: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsSecurityGroupDefaults()
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
			if len(diffs) == 0 {
				return
			}

			for _, change := range diffs {
				t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
			}
		})
	}
}

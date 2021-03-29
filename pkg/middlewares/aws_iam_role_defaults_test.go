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

func TestAwsIamRoleDefaults_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []resource.Resource
		resourcesFromState []resource.Resource
		expected           diff.Changelog
	}{
		{
			"default iam roles when they're not managed by IaC",
			[]resource.Resource{
				&aws.AwsIamRole{
					Id: "AWSServiceRoleForSSO",
				},
				&aws.AwsIamRole{
					Id: "OrganizationAccountAccessRole",
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
			diff.Changelog{},
		},
		{
			"default iam roles when they're managed by IaC",
			[]resource.Resource{
				&aws.AwsIamRole{
					Id: "AWSServiceRoleForSSO",
				},
				&aws.AwsIamRole{
					Id: "OrganizationAccountAccessRole",
				},
				&aws.AwsIamRole{
					Id: "driftctl_assume_role:driftctl_policy.10",
					Tags: map[string]string{
						"test": "value",
					},
				},
			},
			[]resource.Resource{
				&aws.AwsIamRole{
					Id: "AWSServiceRoleForSSO",
				},
				&aws.AwsIamRole{
					Id: "OrganizationAccountAccessRole",
				},
				&aws.AwsIamRole{
					Id:   "driftctl_assume_role:driftctl_policy.10",
					Tags: map[string]string{},
				},
			},
			diff.Changelog{
				{
					Type: "delete",
					Path: []string{"Tags", "2", "test"},
					From: "value",
					To:   nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsIamRoleDefaults()
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

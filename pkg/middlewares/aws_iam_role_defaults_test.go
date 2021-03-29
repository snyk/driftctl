package middlewares

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/r3labs/diff/v2"
)

func TestAwsIamRoleDefaults_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []resource.Resource
		resourcesFromState []resource.Resource
		expected           []resource.Resource
	}{
		{
			"ignore default iam roles when they're managed by IaC",
			[]resource.Resource{
				&aws.AwsIamRole{
					Id: "OrganizationAccountAccessRole:AdministratorAccess",
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
			[]resource.Resource{
				&aws.AwsIamRole{
					Id: "OrganizationAccountAccessRole:AdministratorAccess",
				},
				&aws.AwsRoute{
					Id:           "dummy-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("local"),
				},
			},
		},
		{
			"iam roles when they're managed by IaC",
			[]resource.Resource{
				&aws.AwsIamRole{
					Id: "OrganizationAccountAccessRole:AdministratorAccess",
				},
				&aws.AwsIamRole{
					Id: "driftctl_assume_role:driftctl_policy.10",
				},
				&aws.AwsRoute{
					Id:           "dummy-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("local"),
				},
			},
			[]resource.Resource{
				&aws.AwsIamRole{
					Id: "OrganizationAccountAccessRole:AdministratorAccess",
				},
				&aws.AwsIamRole{
					Id: "driftctl_assume_role:driftctl_policy.10",
				},
				&aws.AwsRoute{
					Id:           "dummy-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("local"),
				},
			},
			[]resource.Resource{
				&aws.AwsIamRole{
					Id: "OrganizationAccountAccessRole:AdministratorAccess",
				},
				&aws.AwsIamRole{
					Id: "driftctl_assume_role:driftctl_policy.10",
				},
				&aws.AwsRoute{
					Id:           "dummy-route",
					RouteTableId: awssdk.String("default-route-table"),
					GatewayId:    awssdk.String("local"),
				},
			},
		},
	}

	differ, err := diff.NewDiffer(diff.SliceOrdering(true))
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsIamRoleDefaults()
			err := m.Execute(&tt.remoteResources, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}
			changelog, err := differ.Diff(tt.expected, tt.remoteResources)
			if err != nil {
				t.Fatal(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
				}
			}
		})
	}
}

package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

func TestAwsDefaults_Execute(t *testing.T) {
	tests := []struct {
		name               string
		remoteResources    []resource.Resource
		resourcesFromState []resource.Resource
		expected           diff.Changelog
	}{
		{
			"default iam roles when they're not managed by IaC",
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "AWSServiceRoleForSSO",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/sso.amazonaws.com",
					},
				},
				&resource.AbstractResource{
					Id:   "OrganizationAccountAccessRole",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/sso.amazonaws.com/",
					},
				},
				&resource.AbstractResource{
					Id:   "terraform-20210408093258091700000001",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/",
					},
				},
				&resource.AbstractResource{
					Id:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
			},
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
				&resource.AbstractResource{
					Id:   "terraform-20210408093258091700000001",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/",
					},
				},
			},
			diff.Changelog{
				{
					Type: "delete",
					Path: []string{"0"},
					From: &resource.AbstractResource{
						Id:   "OrganizationAccountAccessRole",
						Type: aws.AwsIamRoleResourceType,
						Attrs: &resource.Attributes{
							"path": "/not-aws-service-role/sso.amazonaws.com/",
						},
					},
					To: nil,
				},
			},
		},
		{
			"default iam roles when they're managed by IaC",
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "AWSServiceRoleForSSO",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path":        "/aws-service-role/sso.amazonaws.com/",
						"description": "test",
					},
				},
				&resource.AbstractResource{
					Id:   "OrganizationAccountAccessRole",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/sso.amazonaws.com/",
					},
				},
				&resource.AbstractResource{
					Id:   "driftctl_assume_role:driftctl_policy.10",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/",
						"tags": map[string]string{
							"test": "value",
						},
					},
				},
			},
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "AWSServiceRoleForSSO",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/sso.amazonaws.com/",
					},
				},
				&resource.AbstractResource{
					Id:   "OrganizationAccountAccessRole",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/sso.amazonaws.com/",
					},
				},
				&resource.AbstractResource{
					Id:   "driftctl_assume_role:driftctl_policy.10",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/",
						"tags": map[string]string{},
					},
				},
			},
			diff.Changelog{
				{
					Type: diff.DELETE,
					Path: []string{"1", "Attrs", "tags", "test"},
					From: "value",
					To:   nil,
				},
			},
		},
		{
			"ignore default iam role policies when they're not managed by IaC",
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "AWSServiceRoleForSSO",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/sso.amazonaws.com",
					},
				},
				&resource.AbstractResource{
					Id:   "OrganizationAccountAccessRole",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/sso.amazonaws.com",
					},
				},
				&resource.AbstractResource{
					Id:   "AWSServiceRoleForSSO",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role": "AWSServiceRoleForSSO",
					},
				},
				&resource.AbstractResource{
					Id:   "OrganizationAccountAccessRole",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role": "OrganizationAccountAccessRole",
					},
				},
				&resource.AbstractResource{
					Id:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
			},
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "dummy-route",
					Type: aws.AwsRouteResourceType,
					Attrs: &resource.Attributes{
						"route_table_id": "default-route-table",
						"gateway_id":     "local",
					},
				},
			},
			diff.Changelog{
				{
					Type: diff.DELETE,
					Path: []string{"0"},
					From: &resource.AbstractResource{
						Id:   "OrganizationAccountAccessRole",
						Type: aws.AwsIamRoleResourceType,
						Attrs: &resource.Attributes{
							"path": "/not-aws-service-role/sso.amazonaws.com",
						},
					},
					To: nil,
				},
				{
					Type: diff.DELETE,
					Path: []string{"1"},
					From: &resource.AbstractResource{
						Id:   "OrganizationAccountAccessRole",
						Type: aws.AwsIamRolePolicyResourceType,
						Attrs: &resource.Attributes{
							"role": "OrganizationAccountAccessRole",
						},
					},
					To: nil,
				},
			},
		},
		{
			"ignore default iam role policies even when they're managed by IaC",
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "custom-role",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/not-aws-service-role/sso.amazonaws.com",
					},
				},
				&resource.AbstractResource{
					Id:   "OrganizationAccountAccessRole",
					Type: aws.AwsIamRoleResourceType,
					Attrs: &resource.Attributes{
						"path": "/aws-service-role/sso.amazonaws.com",
					},
				},
				&resource.AbstractResource{
					Id:   "driftctl_assume_role:driftctl_policy.10",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role": "custom-role",
					},
				},
				&resource.AbstractResource{
					Id:   "OrganizationAccountAccessRole:AdministratorAccess",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role":        "OrganizationAccountAccessRole",
						"name_prefix": nil,
					},
				},
			},
			[]resource.Resource{
				&resource.AbstractResource{
					Id:   "OrganizationAccountAccessRole:AdministratorAccess",
					Type: aws.AwsIamRolePolicyResourceType,
					Attrs: &resource.Attributes{
						"role":        "OrganizationAccountAccessRole",
						"name_prefix": "tf-",
					},
				},
			},
			diff.Changelog{
				{
					Type: diff.DELETE,
					Path: []string{"0"},
					From: &resource.AbstractResource{
						Id:   "custom-role",
						Type: aws.AwsIamRoleResourceType,
						Attrs: &resource.Attributes{
							"path": "/not-aws-service-role/sso.amazonaws.com",
						},
					},
					To: nil,
				},
				{
					Type: diff.DELETE,
					Path: []string{"1"},
					From: &resource.AbstractResource{
						Id:   "driftctl_assume_role:driftctl_policy.10",
						Type: aws.AwsIamRolePolicyResourceType,
						Attrs: &resource.Attributes{
							"role": "custom-role",
						},
					},
					To: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &AwsDefaults{}
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

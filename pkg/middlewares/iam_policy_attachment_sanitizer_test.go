package middlewares

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func TestIamPolicyAttachmentSanitizer_Execute(t *testing.T) {
	type resources struct {
		RemoteResources    *[]resource.Resource
		ResourcesFromState *[]resource.Resource
	}
	tests := []struct {
		name     string
		args     resources
		expected resources
		wantErr  bool
	}{
		{
			name: "Split users and ReId", args: struct {
				RemoteResources    *[]resource.Resource
				ResourcesFromState *[]resource.Resource
			}{
				RemoteResources: &[]resource.Resource{
					&aws.AwsIamPolicyAttachment{
						Id:        "wrongId",
						PolicyArn: awssdk.String("arn"),
						Users:     []string{"jean", "paul", "pierre"},
					},
					&aws.AwsIamPolicyAttachment{
						Id:        "wrongId2",
						PolicyArn: awssdk.String("thisisarn"),
						Users:     []string{"jean", "paul", "jacques"},
					},
				},
				ResourcesFromState: &[]resource.Resource{
					&aws.AwsIamPolicyAttachment{
						Id:        "wrongId",
						PolicyArn: awssdk.String("fromstatearn"),
						Users:     []string{"jean"},
					},
				},
			},
			expected: struct {
				RemoteResources    *[]resource.Resource
				ResourcesFromState *[]resource.Resource
			}{
				RemoteResources: &[]resource.Resource{
					&aws.AwsIamPolicyAttachment{
						Id:        "jean-arn",
						PolicyArn: awssdk.String("arn"),
						Users:     []string{"jean"},
					},
					&aws.AwsIamPolicyAttachment{
						Id:        "paul-arn",
						PolicyArn: awssdk.String("arn"),
						Users:     []string{"paul"},
					},
					&aws.AwsIamPolicyAttachment{
						Id:        "pierre-arn",
						PolicyArn: awssdk.String("arn"),
						Users:     []string{"pierre"},
					},
					&aws.AwsIamPolicyAttachment{
						Id:        "jean-thisisarn",
						PolicyArn: awssdk.String("thisisarn"),
						Users:     []string{"jean"},
					},
					&aws.AwsIamPolicyAttachment{
						Id:        "paul-thisisarn",
						PolicyArn: awssdk.String("thisisarn"),
						Users:     []string{"paul"},
					},
					&aws.AwsIamPolicyAttachment{
						Id:        "jacques-thisisarn",
						PolicyArn: awssdk.String("thisisarn"),
						Users:     []string{"jacques"},
					},
				},
				ResourcesFromState: &[]resource.Resource{
					&aws.AwsIamPolicyAttachment{
						Id:        "jean-fromstatearn",
						PolicyArn: awssdk.String("fromstatearn"),
						Users:     []string{"jean"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Split Roles and ReId", args: struct {
				RemoteResources    *[]resource.Resource
				ResourcesFromState *[]resource.Resource
			}{
				RemoteResources: &[]resource.Resource{
					&aws.AwsIamPolicyAttachment{
						Id:        "wrongId",
						PolicyArn: awssdk.String("arn"),
						Roles:     []string{"role1", "role2", "pierre"},
					},
					&aws.AwsIamPolicyAttachment{
						Id:        "wrongId2",
						PolicyArn: awssdk.String("thisisarn"),
						Roles:     []string{"role1", "role2", "role3"},
					},
				},
				ResourcesFromState: &[]resource.Resource{
					&aws.AwsIamPolicyAttachment{
						Id:        "wrongId",
						PolicyArn: awssdk.String("fromstatearn"),
						Roles:     []string{"role1"},
					},
				},
			},
			expected: struct {
				RemoteResources    *[]resource.Resource
				ResourcesFromState *[]resource.Resource
			}{
				RemoteResources: &[]resource.Resource{
					&aws.AwsIamPolicyAttachment{
						Id:        "role1-arn",
						PolicyArn: awssdk.String("arn"),
						Roles:     []string{"role1"},
					},
					&aws.AwsIamPolicyAttachment{
						Id:        "role2-arn",
						PolicyArn: awssdk.String("arn"),
						Roles:     []string{"role2"},
					},
					&aws.AwsIamPolicyAttachment{
						Id:        "pierre-arn",
						PolicyArn: awssdk.String("arn"),
						Roles:     []string{"pierre"},
					},
					&aws.AwsIamPolicyAttachment{
						Id:        "role1-thisisarn",
						PolicyArn: awssdk.String("thisisarn"),
						Roles:     []string{"role1"},
					},
					&aws.AwsIamPolicyAttachment{
						Id:        "role2-thisisarn",
						PolicyArn: awssdk.String("thisisarn"),
						Roles:     []string{"role2"},
					},
					&aws.AwsIamPolicyAttachment{
						Id:        "role3-thisisarn",
						PolicyArn: awssdk.String("thisisarn"),
						Roles:     []string{"role3"},
					},
				},
				ResourcesFromState: &[]resource.Resource{
					&aws.AwsIamPolicyAttachment{
						Id:        "role1-fromstatearn",
						PolicyArn: awssdk.String("fromstatearn"),
						Roles:     []string{"role1"},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := IamPolicyAttachmentSanitizer{}
			if err := m.Execute(tt.args.RemoteResources, tt.args.ResourcesFromState); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			changelog, err := diff.Diff(tt.args, tt.expected)
			if err != nil {
				t.Error(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
				}
			}
		})
	}
}

package middlewares

import (
	"strings"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func TestIamPolicyAttachmentExpander_Execute(t *testing.T) {
	type resources struct {
		RemoteResources    *[]*resource.Resource
		ResourcesFromState *[]*resource.Resource
	}
	tests := []struct {
		name     string
		args     resources
		mocks    func(*terraform.MockResourceFactory)
		expected resources
		wantErr  bool
	}{
		{
			name: "Split users and ReId",
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"jean-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"users":      []interface{}{"jean"},
					},
				).Once().Return(&resource.Resource{
					Id:   "jean-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"paul-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"users":      []interface{}{"paul"},
					},
				).Once().Return(&resource.Resource{
					Id:   "paul-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"pierre-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"users":      []interface{}{"pierre"},
					},
				).Once().Return(&resource.Resource{
					Id:   "pierre-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"jean-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"users":      []interface{}{"jean"},
					},
				).Once().Return(&resource.Resource{
					Id:   "jean-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"paul-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"users":      []interface{}{"paul"},
					},
				).Once().Return(&resource.Resource{
					Id:   "paul-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"jacques-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"users":      []interface{}{"jacques"},
					},
				).Once().Return(&resource.Resource{
					Id:   "jacques-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"jean-fromstatearn",
					map[string]interface{}{
						"policy_arn": "fromstatearn",
						"users":      []interface{}{"jean"},
					},
				).Once().Return(&resource.Resource{
					Id:   "jean-fromstatearn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
			},
			args: struct {
				RemoteResources    *[]*resource.Resource
				ResourcesFromState *[]*resource.Resource
			}{
				RemoteResources: &[]*resource.Resource{
					{
						Id:   "wrongId",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "arn",
							"users":      []interface{}{"jean", "paul", "pierre"},
						},
					},
					{
						Id:   "wrongId2",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "thisisarn",
							"users":      []interface{}{"jean", "paul", "jacques"},
						},
					},
				},
				ResourcesFromState: &[]*resource.Resource{
					{
						Id:   "wrongId",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "fromstatearn",
							"users":      []interface{}{"jean"},
						},
					},
				},
			},
			expected: struct {
				RemoteResources    *[]*resource.Resource
				ResourcesFromState *[]*resource.Resource
			}{
				RemoteResources: &[]*resource.Resource{
					{
						Id:   "jean-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						Id:   "paul-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						Id:   "pierre-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						Id:   "jean-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						Id:   "paul-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						Id:   "jacques-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
				},
				ResourcesFromState: &[]*resource.Resource{
					{
						Id:   "jean-fromstatearn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Split Roles and ReId",
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role1-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"roles":      []interface{}{"role1"},
					},
				).Once().Return(&resource.Resource{
					Id:   "role1-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role2-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"roles":      []interface{}{"role2"},
					},
				).Once().Return(&resource.Resource{
					Id:   "role2-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"pierre-arn",
					map[string]interface{}{
						"policy_arn": "arn",
						"roles":      []interface{}{"pierre"},
					},
				).Once().Return(&resource.Resource{
					Id:   "pierre-arn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role1-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"roles":      []interface{}{"role1"},
					},
				).Once().Return(&resource.Resource{
					Id:   "role1-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role2-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"roles":      []interface{}{"role2"},
					},
				).Once().Return(&resource.Resource{
					Id:   "role2-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role3-thisisarn",
					map[string]interface{}{
						"policy_arn": "thisisarn",
						"roles":      []interface{}{"role3"},
					},
				).Once().Return(&resource.Resource{
					Id:   "role3-thisisarn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
				factory.On(
					"CreateAbstractResource",
					aws.AwsIamPolicyAttachmentResourceType,
					"role1-fromstatearn",
					map[string]interface{}{
						"policy_arn": "fromstatearn",
						"roles":      []interface{}{"role1"},
					},
				).Once().Return(&resource.Resource{
					Id:   "role1-fromstatearn",
					Type: aws.AwsIamPolicyAttachmentResourceType,
				})
			},
			args: struct {
				RemoteResources    *[]*resource.Resource
				ResourcesFromState *[]*resource.Resource
			}{
				RemoteResources: &[]*resource.Resource{
					{
						Id:   "wrongId",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "arn",
							"roles":      []interface{}{"role1", "role2", "pierre"},
						},
					},
					{
						Id:   "wrongId2",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "thisisarn",
							"roles":      []interface{}{"role1", "role2", "role3"},
						},
					},
				},
				ResourcesFromState: &[]*resource.Resource{
					{
						Id:   "wrongId",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "fromstatearn",
							"roles":      []interface{}{"role1"},
						},
					},
				},
			},
			expected: struct {
				RemoteResources    *[]*resource.Resource
				ResourcesFromState *[]*resource.Resource
			}{
				RemoteResources: &[]*resource.Resource{
					{
						Id:   "role1-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						Id:   "role2-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						Id:   "pierre-arn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						Id:   "role1-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						Id:   "role2-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
					{
						Id:   "role3-thisisarn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
				},
				ResourcesFromState: &[]*resource.Resource{
					{
						Id:   "role1-fromstatearn",
						Type: aws.AwsIamPolicyAttachmentResourceType,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			factory := &terraform.MockResourceFactory{}
			if tt.mocks != nil {
				tt.mocks(factory)
			}

			m := NewIamPolicyAttachmentExpander(factory)
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

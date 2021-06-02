package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/r3labs/diff/v2"
)

func TestIamPolicyAttachmentTransformer_Execute(t *testing.T) {
	type argRes struct {
		RemoteResources    *[]resource.Resource
		ResourcesFromState *[]resource.Resource
	}
	tests := []struct {
		name     string
		args     argRes
		expected argRes
		mocks    func(factory *terraform.MockResourceFactory)
	}{
		{
			name: "transform user_policy_attachment",
			args: argRes{
				RemoteResources: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "id1",
						Type: aws.AwsIamUserPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "policy_arn1",
							"user":       "user1",
						},
					},
				},
				ResourcesFromState: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "id2",
						Type: aws.AwsIamUserPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "policy_arn2",
							"user":       "user2",
						},
					},
				},
			},
			expected: argRes{
				RemoteResources: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "id1",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"id":         "id1",
							"policy_arn": "policy_arn1",
							"users":      []interface{}{"user1"},
							"groups":     []interface{}{},
							"roles":      []interface{}{},
						},
					},
				},
				ResourcesFromState: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "id2",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"id":         "id2",
							"policy_arn": "policy_arn2",
							"users":      []interface{}{"user2"},
							"groups":     []interface{}{},
							"roles":      []interface{}{},
						},
					},
				},
			},
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On("CreateAbstractResource", aws.AwsIamPolicyAttachmentResourceType, "id1", map[string]interface{}{
					"id":         "id1",
					"policy_arn": "policy_arn1",
					"users":      []interface{}{"user1"},
					"groups":     []interface{}{},
					"roles":      []interface{}{},
				}).Once().Return(&resource.AbstractResource{
					Id:   "id1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"id":         "id1",
						"policy_arn": "policy_arn1",
						"users":      []interface{}{"user1"},
						"groups":     []interface{}{},
						"roles":      []interface{}{},
					},
				}, nil)
				factory.On("CreateAbstractResource", aws.AwsIamPolicyAttachmentResourceType, "id2", map[string]interface{}{
					"id":         "id2",
					"policy_arn": "policy_arn2",
					"users":      []interface{}{"user2"},
					"groups":     []interface{}{},
					"roles":      []interface{}{},
				}).Once().Return(&resource.AbstractResource{
					Id:   "id2",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"id":         "id2",
						"policy_arn": "policy_arn2",
						"users":      []interface{}{"user2"},
						"groups":     []interface{}{},
						"roles":      []interface{}{},
					},
				}, nil)
			},
		},
		{
			name: "transform role_policy_attachment",
			args: argRes{
				RemoteResources: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "id1",
						Type: aws.AwsIamRolePolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "policy_arn1",
							"role":       "role1",
						},
					},
				},
				ResourcesFromState: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "id2",
						Type: aws.AwsIamRolePolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"policy_arn": "policy_arn2",
							"role":       "role2",
						},
					},
				},
			},
			expected: argRes{
				RemoteResources: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "id1",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"id":         "id1",
							"policy_arn": "policy_arn1",
							"users":      []interface{}{},
							"groups":     []interface{}{},
							"roles":      []interface{}{"role1"},
						},
					},
				},
				ResourcesFromState: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "id2",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"id":         "id2",
							"policy_arn": "policy_arn2",
							"users":      []interface{}{},
							"groups":     []interface{}{},
							"roles":      []interface{}{"role2"},
						},
					},
				},
			},
			mocks: func(factory *terraform.MockResourceFactory) {
				factory.On("CreateAbstractResource", aws.AwsIamPolicyAttachmentResourceType, "id1", map[string]interface{}{
					"id":         "id1",
					"policy_arn": "policy_arn1",
					"users":      []interface{}{},
					"groups":     []interface{}{},
					"roles":      []interface{}{"role1"},
				}).Once().Return(&resource.AbstractResource{
					Id:   "id1",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"id":         "id1",
						"policy_arn": "policy_arn1",
						"users":      []interface{}{},
						"groups":     []interface{}{},
						"roles":      []interface{}{"role1"},
					},
				}, nil)
				factory.On("CreateAbstractResource", aws.AwsIamPolicyAttachmentResourceType, "id2", map[string]interface{}{
					"id":         "id2",
					"policy_arn": "policy_arn2",
					"users":      []interface{}{},
					"groups":     []interface{}{},
					"roles":      []interface{}{"role2"},
				}).Once().Return(&resource.AbstractResource{
					Id:   "id2",
					Type: aws.AwsIamPolicyAttachmentResourceType,
					Attrs: &resource.Attributes{
						"id":         "id2",
						"policy_arn": "policy_arn2",
						"users":      []interface{}{},
						"groups":     []interface{}{},
						"roles":      []interface{}{"role2"},
					},
				}, nil)
			},
		},
		{
			name: "transform nothing",
			args: argRes{
				RemoteResources: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "id1",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"id":         "id1",
							"policy_arn": "policy_arn1",
							"users":      []interface{}{},
							"groups":     []interface{}{},
							"roles":      []interface{}{"role1"},
						},
					},
				},
				ResourcesFromState: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "id2",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"id":         "id2",
							"policy_arn": "policy_arn2",
							"users":      []interface{}{},
							"groups":     []interface{}{},
							"roles":      []interface{}{"role2"},
						},
					},
				},
			},
			expected: argRes{
				RemoteResources: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "id1",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"id":         "id1",
							"policy_arn": "policy_arn1",
							"users":      []interface{}{},
							"groups":     []interface{}{},
							"roles":      []interface{}{"role1"},
						},
					},
				},
				ResourcesFromState: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "id2",
						Type: aws.AwsIamPolicyAttachmentResourceType,
						Attrs: &resource.Attributes{
							"id":         "id2",
							"policy_arn": "policy_arn2",
							"users":      []interface{}{},
							"groups":     []interface{}{},
							"roles":      []interface{}{"role2"},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &terraform.MockResourceFactory{}
			if tt.mocks != nil {
				tt.mocks(factory)
			}

			repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
			aws.InitResourcesMetadata(repo)
			m := IamPolicyAttachmentTransformer{
				resourceFactory: factory,
			}
			if err := m.Execute(tt.args.RemoteResources, tt.args.ResourcesFromState); err != nil {
				t.Errorf(err.Error())
			}

			changelog, err := diff.Diff(tt.expected, tt.args)
			if err != nil {
				t.Errorf(err.Error())
			}

			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
				}
			}
		})
	}
}

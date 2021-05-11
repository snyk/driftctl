package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/stretchr/testify/mock"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/mocks"
)

func Test_IAMRepository_ListAllAccessKeys(t *testing.T) {
	tests := []struct {
		name    string
		users   []*iam.User
		mocks   func(client *mocks.FakeIAM)
		want    []*iam.AccessKeyMetadata
		wantErr error
	}{
		{
			name: "List only access keys with multiple pages",
			users: []*iam.User{
				{
					UserName: aws.String("test-driftctl"),
				},
				{
					UserName: aws.String("test-driftctl2"),
				},
			},
			mocks: func(client *mocks.FakeIAM) {

				client.On("ListAccessKeysPages",
					&iam.ListAccessKeysInput{
						UserName: aws.String("test-driftctl"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListAccessKeysOutput, lastPage bool) bool) bool {
						callback(&iam.ListAccessKeysOutput{AccessKeyMetadata: []*iam.AccessKeyMetadata{
							{
								AccessKeyId: aws.String("AKIA5QYBVVD223VWU32A"),
								UserName:    aws.String("test-driftctl"),
							},
						}}, false)
						callback(&iam.ListAccessKeysOutput{AccessKeyMetadata: []*iam.AccessKeyMetadata{
							{
								AccessKeyId: aws.String("AKIA5QYBVVD2QYI36UZP"),
								UserName:    aws.String("test-driftctl"),
							},
						}}, true)
						return true
					})).Return(nil)
				client.On("ListAccessKeysPages",
					&iam.ListAccessKeysInput{
						UserName: aws.String("test-driftctl2"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListAccessKeysOutput, lastPage bool) bool) bool {
						callback(&iam.ListAccessKeysOutput{AccessKeyMetadata: []*iam.AccessKeyMetadata{
							{
								AccessKeyId: aws.String("AKIA5QYBVVD26EJME25D"),
								UserName:    aws.String("test-driftctl2"),
							},
						}}, false)
						callback(&iam.ListAccessKeysOutput{AccessKeyMetadata: []*iam.AccessKeyMetadata{
							{
								AccessKeyId: aws.String("AKIA5QYBVVD2SWDFVVMG"),
								UserName:    aws.String("test-driftctl2"),
							},
						}}, true)
						return true
					})).Return(nil)
			},
			want: []*iam.AccessKeyMetadata{
				{
					AccessKeyId: aws.String("AKIA5QYBVVD223VWU32A"),
					UserName:    aws.String("test-driftctl"),
				},
				{
					AccessKeyId: aws.String("AKIA5QYBVVD2QYI36UZP"),
					UserName:    aws.String("test-driftctl"),
				},
				{
					AccessKeyId: aws.String("AKIA5QYBVVD223VWU32A"),
					UserName:    aws.String("test-driftctl"),
				},
				{
					AccessKeyId: aws.String("AKIA5QYBVVD2QYI36UZP"),
					UserName:    aws.String("test-driftctl"),
				},
				{
					AccessKeyId: aws.String("AKIA5QYBVVD26EJME25D"),
					UserName:    aws.String("test-driftctl2"),
				},
				{
					AccessKeyId: aws.String("AKIA5QYBVVD2SWDFVVMG"),
					UserName:    aws.String("test-driftctl2"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mocks.FakeIAM{}
			tt.mocks(client)
			r := &iamRepository{
				client: client,
			}
			got, err := r.ListAllAccessKeys(tt.users)
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %v -> %v", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}
func Test_IAMRepository_ListAllRoles(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *mocks.FakeIAM)
		want    []*iam.Role
		wantErr error
	}{
		{
			name: "List only roles with multiple pages",
			mocks: func(client *mocks.FakeIAM) {

				client.On("ListRolesPages",
					&iam.ListRolesInput{},
					mock.MatchedBy(func(callback func(res *iam.ListRolesOutput, lastPage bool) bool) bool {
						callback(&iam.ListRolesOutput{Roles: []*iam.Role{
							{
								RoleName: aws.String("test-driftctl"),
							},
							{
								RoleName: aws.String("test-driftctl2"),
							},
						}}, false)
						callback(&iam.ListRolesOutput{Roles: []*iam.Role{
							{
								RoleName: aws.String("test-driftctl3"),
							},
							{
								RoleName: aws.String("test-driftctl4"),
							},
						}}, true)
						return true
					})).Return(nil)
			},
			want: []*iam.Role{
				{
					RoleName: aws.String("test-driftctl"),
				},
				{
					RoleName: aws.String("test-driftctl2"),
				},
				{
					RoleName: aws.String("test-driftctl3"),
				},
				{
					RoleName: aws.String("test-driftctl4"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mocks.FakeIAM{}
			tt.mocks(client)
			r := &iamRepository{
				client: client,
			}
			got, err := r.ListAllRoles()
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %v -> %v", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_IAMRepository_ListAllRolePolicies(t *testing.T) {
	tests := []struct {
		name    string
		roles   []*iam.Role
		mocks   func(client *mocks.FakeIAM)
		want    []string
		wantErr error
	}{
		{
			name: "List only role policies with multiple pages",
			roles: []*iam.Role{
				{
					RoleName: aws.String("test_role_0"),
				},
				{
					RoleName: aws.String("test_role_1"),
				},
			},
			mocks: func(client *mocks.FakeIAM) {
				firstMockCalled := false
				client.On("ListRolePoliciesPages",
					&iam.ListRolePoliciesInput{
						RoleName: aws.String("test_role_0"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListRolePoliciesOutput, lastPage bool) bool) bool {
						if firstMockCalled {
							return false
						}
						callback(&iam.ListRolePoliciesOutput{
							PolicyNames: []*string{
								aws.String("policy-role0-0"),
								aws.String("policy-role0-1"),
							},
						}, false)
						callback(&iam.ListRolePoliciesOutput{
							PolicyNames: []*string{
								aws.String("policy-role0-2"),
							},
						}, true)
						firstMockCalled = true
						return true
					})).Once().Return(nil)
				client.On("ListRolePoliciesPages",
					&iam.ListRolePoliciesInput{
						RoleName: aws.String("test_role_1"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListRolePoliciesOutput, lastPage bool) bool) bool {
						callback(&iam.ListRolePoliciesOutput{
							PolicyNames: []*string{
								aws.String("policy-role1-0"),
								aws.String("policy-role1-1"),
							},
						}, false)
						callback(&iam.ListRolePoliciesOutput{
							PolicyNames: []*string{
								aws.String("policy-role1-2"),
							},
						}, true)
						return true
					})).Once().Return(nil)
			},
			want: []string{
				*aws.String("test_role_0:policy-role0-0"),
				*aws.String("test_role_0:policy-role0-1"),
				*aws.String("test_role_0:policy-role0-2"),
				*aws.String("test_role_1:policy-role1-0"),
				*aws.String("test_role_1:policy-role1-1"),
				*aws.String("test_role_1:policy-role1-2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mocks.FakeIAM{}
			tt.mocks(client)
			r := &iamRepository{
				client: client,
			}
			got, err := r.ListAllRolePolicies(tt.roles)
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %v -> %v", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}
func Test_IAMRepository_ListAllUserPolicyAttachments(t *testing.T) {
	tests := []struct {
		name    string
		users   []*iam.User
		mocks   func(client *mocks.FakeIAM)
		want    []*AttachedUserPolicy
		wantErr error
	}{
		{
			name: "List only user policy attachments with multiple pages",
			users: []*iam.User{
				{
					UserName: aws.String("loadbalancer"),
				},
				{
					UserName: aws.String("loadbalancer2"),
				},
			},
			mocks: func(client *mocks.FakeIAM) {

				client.On("ListAttachedUserPoliciesPages",
					&iam.ListAttachedUserPoliciesInput{
						UserName: aws.String("loadbalancer"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListAttachedUserPoliciesOutput, lastPage bool) bool) bool {
						callback(&iam.ListAttachedUserPoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test"),
								PolicyName: aws.String("test-attach"),
							},
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test2"),
								PolicyName: aws.String("test-attach2"),
							},
						}}, false)
						callback(&iam.ListAttachedUserPoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test3"),
								PolicyName: aws.String("test-attach3"),
							},
						}}, true)
						return true
					})).Return(nil).Once()

				client.On("ListAttachedUserPoliciesPages",
					&iam.ListAttachedUserPoliciesInput{
						UserName: aws.String("loadbalancer2"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListAttachedUserPoliciesOutput, lastPage bool) bool) bool {
						callback(&iam.ListAttachedUserPoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test"),
								PolicyName: aws.String("test-attach"),
							},
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test2"),
								PolicyName: aws.String("test-attach2"),
							},
						}}, false)
						callback(&iam.ListAttachedUserPoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test3"),
								PolicyName: aws.String("test-attach3"),
							},
						}}, true)
						return true
					})).Return(nil).Once()
			},

			want: []*AttachedUserPolicy{
				{
					iam.AttachedPolicy{
						PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test"),
						PolicyName: aws.String("test-attach"),
					},
					*aws.String("loadbalancer"),
				},
				{
					iam.AttachedPolicy{
						PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test2"),
						PolicyName: aws.String("test-attach2"),
					},
					*aws.String("loadbalancer"),
				},
				{
					iam.AttachedPolicy{
						PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test3"),
						PolicyName: aws.String("test-attach3"),
					},
					*aws.String("loadbalancer"),
				},
				{
					iam.AttachedPolicy{
						PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test"),
						PolicyName: aws.String("test-attach"),
					},
					*aws.String("loadbalancer2"),
				},
				{
					iam.AttachedPolicy{
						PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test2"),
						PolicyName: aws.String("test-attach2"),
					},
					*aws.String("loadbalancer2"),
				},
				{
					iam.AttachedPolicy{
						PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test3"),
						PolicyName: aws.String("test-attach3"),
					},
					*aws.String("loadbalancer2"),
				},
				{
					iam.AttachedPolicy{
						PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test"),
						PolicyName: aws.String("test-attach"),
					},
					*aws.String("loadbalancer2"),
				},
				{
					iam.AttachedPolicy{
						PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test2"),
						PolicyName: aws.String("test-attach2"),
					},
					*aws.String("loadbalancer2"),
				},
				{
					iam.AttachedPolicy{
						PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test3"),
						PolicyName: aws.String("test-attach3"),
					},
					*aws.String("loadbalancer2"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mocks.FakeIAM{}
			tt.mocks(client)
			r := &iamRepository{
				client: client,
			}
			got, err := r.ListAllUserPolicyAttachments(tt.users)
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %v -> %v", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_IAMRepository_ListAllUserPolicies(t *testing.T) {
	tests := []struct {
		name    string
		users   []*iam.User
		mocks   func(client *mocks.FakeIAM)
		want    []string
		wantErr error
	}{
		{
			name: "List only user policies with multiple pages",
			users: []*iam.User{
				{
					UserName: aws.String("loadbalancer"),
				},
				{
					UserName: aws.String("loadbalancer2"),
				},
			},
			mocks: func(client *mocks.FakeIAM) {

				client.On("ListUserPoliciesPages",
					&iam.ListUserPoliciesInput{
						UserName: aws.String("loadbalancer"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListUserPoliciesOutput, lastPage bool) bool) bool {
						callback(&iam.ListUserPoliciesOutput{PolicyNames: []*string{
							aws.String("test"),
							aws.String("test2"),
							aws.String("test3"),
						}}, false)
						callback(&iam.ListUserPoliciesOutput{PolicyNames: []*string{
							aws.String("test4"),
						}}, true)
						return true
					})).Return(nil).Once()

				client.On("ListUserPoliciesPages",
					&iam.ListUserPoliciesInput{
						UserName: aws.String("loadbalancer2"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListUserPoliciesOutput, lastPage bool) bool) bool {
						callback(&iam.ListUserPoliciesOutput{PolicyNames: []*string{
							aws.String("test2"),
							aws.String("test22"),
							aws.String("test23"),
						}}, false)
						callback(&iam.ListUserPoliciesOutput{PolicyNames: []*string{
							aws.String("test24"),
						}}, true)
						return true
					})).Return(nil).Once()
			},
			want: []string{
				*aws.String("loadbalancer:test"),
				*aws.String("loadbalancer:test2"),
				*aws.String("loadbalancer:test3"),
				*aws.String("loadbalancer:test4"),
				*aws.String("loadbalancer2:test"),
				*aws.String("loadbalancer2:test2"),
				*aws.String("loadbalancer2:test3"),
				*aws.String("loadbalancer2:test4"),
				*aws.String("loadbalancer2:test2"),
				*aws.String("loadbalancer2:test22"),
				*aws.String("loadbalancer2:test23"),
				*aws.String("loadbalancer2:test24"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mocks.FakeIAM{}
			tt.mocks(client)
			r := &iamRepository{
				client: client,
			}
			got, err := r.ListAllUserPolicies(tt.users)
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %v -> %v", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/mocks"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
)

func TestIamRolePolicyAttachmentSupplier_Resources(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(client *mocks.FakeIAM)
		err     error
	}{
		{
			test:    "iam multiples roles multiple policies",
			dirName: "iam_role_policy_attachment_multiple",
			mocks: func(client *mocks.FakeIAM) {
				client.On("ListRolesPages",
					&iam.ListRolesInput{},
					mock.MatchedBy(func(callback func(res *iam.ListRolesOutput, lastPage bool) bool) bool {
						callback(&iam.ListRolesOutput{Roles: []*iam.Role{
							{
								RoleName: aws.String("test-role"),
							},
							{
								RoleName: aws.String("test-role2"),
							},
						}}, true)
						return true
					})).Return(nil).Once()

				shouldSkipfirst := false
				shouldSkipSecond := false

				client.On("ListAttachedRolePoliciesPages",
					&iam.ListAttachedRolePoliciesInput{
						RoleName: aws.String("test-role"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool) bool {
						if shouldSkipfirst {
							return false
						}
						callback(&iam.ListAttachedRolePoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test-policy"),
								PolicyName: aws.String("policy"),
							},
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test-policy2"),
								PolicyName: aws.String("policy2"),
							},
						}}, false)
						callback(&iam.ListAttachedRolePoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test-policy3"),
								PolicyName: aws.String("policy3"),
							},
						}}, true)
						shouldSkipfirst = true
						return true
					})).Return(nil).Once()

				client.On("ListAttachedRolePoliciesPages",
					&iam.ListAttachedRolePoliciesInput{
						RoleName: aws.String("test-role2"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool) bool {
						if shouldSkipSecond {
							return false
						}
						callback(&iam.ListAttachedRolePoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test-policy"),
								PolicyName: aws.String("policy"),
							},
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test-policy2"),
								PolicyName: aws.String("policy2"),
							},
						}}, false)
						callback(&iam.ListAttachedRolePoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test-policy3"),
								PolicyName: aws.String("policy3"),
							},
						}}, true)
						shouldSkipSecond = true
						return true
					})).Return(nil).Once()
			},
			err: nil,
		},
		{
			test:    "check that we ignore policy for ignored roles",
			dirName: "iam_role_policy_attachment_for_ignored_roles",
			mocks: func(client *mocks.FakeIAM) {
				client.On("ListRolesPages",
					&iam.ListRolesInput{},
					mock.MatchedBy(func(callback func(res *iam.ListRolesOutput, lastPage bool) bool) bool {
						callback(&iam.ListRolesOutput{Roles: []*iam.Role{
							{
								RoleName: aws.String("AWSServiceRoleForOrganizations"),
							},
							{
								RoleName: aws.String("AWSServiceRoleForSupport"),
							},
							{
								RoleName: aws.String("AWSServiceRoleForTrustedAdvisor"),
							},
						}}, true)
						return true
					})).Return(nil)
			},
			err: nil,
		},
		{
			test:    "Cannot list roles",
			dirName: "iam_role_policy_attachment_for_ignored_roles",
			mocks: func(client *mocks.FakeIAM) {
				client.On("ListRolesPages",
					&iam.ListRolesInput{},
					mock.MatchedBy(func(callback func(res *iam.ListRolesOutput, lastPage bool) bool) bool {
						callback(&iam.ListRolesOutput{Roles: []*iam.Role{}}, true)
						return true
					})).Return(awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamRolePolicyAttachmentResourceType, resourceaws.AwsIamRoleResourceType),
		},
		{
			test:    "Cannot list roles policies",
			dirName: "iam_role_policy_attachment_for_ignored_roles",
			mocks: func(client *mocks.FakeIAM) {
				client.On("ListRolesPages",
					&iam.ListRolesInput{},
					mock.MatchedBy(func(callback func(res *iam.ListRolesOutput, lastPage bool) bool) bool {
						callback(&iam.ListRolesOutput{Roles: []*iam.Role{
							{
								RoleName: aws.String("test-role"),
							},
							{
								RoleName: aws.String("test-role2"),
							},
						}}, true)
						return true
					})).Return(nil).Once()
				client.On("ListAttachedRolePoliciesPages",
					mock.Anything,
					mock.MatchedBy(func(callback func(res *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool) bool {
						return true
					})).Return(awserr.NewRequestFailure(nil, 403, "")).Once()
			},
			err: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamRolePolicyAttachmentResourceType, resourceaws.AwsIamRolePolicyResourceType),
		},
	}
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			providerLibrary.AddProvider(terraform.AWS, provider)
			supplierLibrary.AddSupplier(NewIamRolePolicyAttachmentSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeIam := mocks.FakeIAM{}
			c.mocks(&fakeIam)

			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewIamRolePolicyAttachmentDeserializer()
			s := &IamRolePolicyAttachmentSupplier{
				provider,
				deserializer,
				&fakeIam,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 1)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, awsdeserializer.NewIamPolicyAttachmentDeserializer(), shouldUpdate, t)
		})
	}
}

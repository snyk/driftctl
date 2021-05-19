package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
)

func TestIamRolePolicyAttachmentSupplier_Resources(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockIAMRepository)
		err     error
	}{
		{
			test:    "no iam role policy",
			dirName: "iam_role_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				roles := []*iam.Role{
					{
						RoleName: aws.String("test-role"),
					},
				}
				repo.On("ListAllRoles").Return(roles, nil)
				repo.On("ListAllRolePolicyAttachments", roles).Return([]*repository.AttachedRolePolicy{}, nil)
			},
			err: nil,
		},
		{
			test:    "iam multiples roles multiple policies",
			dirName: "iam_role_policy_attachment_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				roles := []*iam.Role{
					{
						RoleName: aws.String("test-role"),
					},
					{
						RoleName: aws.String("test-role2"),
					},
				}
				repo.On("ListAllRoles").Return(roles, nil)
				repo.On("ListAllRolePolicyAttachments", roles).Return([]*repository.AttachedRolePolicy{
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test-policy"),
							PolicyName: aws.String("policy"),
						},
						RoleName: *aws.String("test-role"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test-policy2"),
							PolicyName: aws.String("policy2"),
						},
						RoleName: *aws.String("test-role"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test-policy3"),
							PolicyName: aws.String("policy3"),
						},
						RoleName: *aws.String("test-role"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test-policy"),
							PolicyName: aws.String("policy"),
						},
						RoleName: *aws.String("test-role2"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test-policy2"),
							PolicyName: aws.String("policy2"),
						},
						RoleName: *aws.String("test-role2"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test-policy3"),
							PolicyName: aws.String("policy3"),
						},
						RoleName: *aws.String("test-role2"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "Cannot list roles",
			dirName: "iam_role_policy_attachment_for_ignored_roles",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Once().Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamRolePolicyAttachmentResourceType, resourceaws.AwsIamRoleResourceType),
		},
		{
			test:    "Cannot list roles policy attachment",
			dirName: "iam_role_policy_attachment_for_ignored_roles",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Once().Return([]*iam.Role{}, nil)
				repo.On("ListAllRolePolicyAttachments", mock.Anything).Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamRolePolicyAttachmentResourceType),
		},
	}
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := InitTestAwsProvider(providerLibrary)
			if err != nil {
				t.Fatal(err)
			}
			supplierLibrary.AddSupplier(NewIamRolePolicyAttachmentSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeIam := repository.MockIAMRepository{}
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

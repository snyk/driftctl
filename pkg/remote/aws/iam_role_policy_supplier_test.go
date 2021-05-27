package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/cloudskiff/driftctl/test/mocks"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
)

func TestIamRolePolicySupplier_Resources(t *testing.T) {

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
						RoleName: aws.String("test_role"),
					},
				}
				repo.On("ListAllRoles").Return(roles, nil)
				repo.On("ListAllRolePolicies", roles).Return([]string{}, nil)
			},
			err: nil,
		},
		{
			test:    "multiples roles with inline policies",
			dirName: "iam_role_policy_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				roles := []*iam.Role{
					{
						RoleName: aws.String("test_role_0"),
					},
					{
						RoleName: aws.String("test_role_1"),
					},
				}
				repo.On("ListAllRoles").Return(roles, nil)
				repo.On("ListAllRolePolicies", roles).Return([]string{
					*aws.String("test_role_0:policy-role0-0"),
					*aws.String("test_role_0:policy-role0-1"),
					*aws.String("test_role_0:policy-role0-2"),
					*aws.String("test_role_1:policy-role1-0"),
					*aws.String("test_role_1:policy-role1-1"),
					*aws.String("test_role_1:policy-role1-2"),
				}, nil).Once()
			},
			err: nil,
		},
		{
			test:    "Cannot list roles",
			dirName: "iam_role_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Once().Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamRolePolicyResourceType, resourceaws.AwsIamRoleResourceType),
		},
		{
			test:    "cannot list role policy",
			dirName: "iam_role_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Once().Return([]*iam.Role{}, nil)
				repo.On("ListAllRolePolicies", mock.Anything).Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamRolePolicyResourceType),
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
			supplierLibrary.AddSupplier(NewIamRolePolicySupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeIam := repository.MockIAMRepository{}
			c.mocks(&fakeIam)

			provider := mocks.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewIamRolePolicyDeserializer()
			s := &IamRolePolicySupplier{
				provider,
				deserializer,
				&fakeIam,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}

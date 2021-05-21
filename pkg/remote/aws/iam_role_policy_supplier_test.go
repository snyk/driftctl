package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	awstest "github.com/cloudskiff/driftctl/test/aws"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
)

func TestIamRolePolicySupplier_Resources(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(client *awstest.MockFakeIAM)
		err     error
	}{
		{
			test:    "multiples roles without any inline policies",
			dirName: "iam_role_policy_empty",
			mocks: func(client *awstest.MockFakeIAM) {
				client.On("ListRolesPages",
					&iam.ListRolesInput{},
					mock.MatchedBy(func(callback func(res *iam.ListRolesOutput, lastPage bool) bool) bool {
						callback(&iam.ListRolesOutput{Roles: []*iam.Role{
							{
								RoleName: aws.String("test_role_0"),
							},
							{
								RoleName: aws.String("test_role_1"),
							},
						}}, true)
						return true
					})).Return(nil)
				client.On("ListRolePoliciesPages",
					&iam.ListRolePoliciesInput{
						RoleName: aws.String("test_role_0"),
					},
					mock.Anything,
				).Return(nil)
				client.On("ListRolePoliciesPages",
					&iam.ListRolePoliciesInput{
						RoleName: aws.String("test_role_1"),
					},
					mock.Anything,
				).Return(nil)
			},
			err: nil,
		},
		{
			test:    "iam multiples roles with inline policies",
			dirName: "iam_role_policy_multiple",
			mocks: func(client *awstest.MockFakeIAM) {
				client.On("ListRolesPages",
					&iam.ListRolesInput{},
					mock.MatchedBy(func(callback func(res *iam.ListRolesOutput, lastPage bool) bool) bool {
						callback(&iam.ListRolesOutput{Roles: []*iam.Role{
							{
								RoleName: aws.String("test_role_0"),
							},
							{
								RoleName: aws.String("test_role_1"),
							},
						}}, true)
						return true
					})).Once().Return(nil)
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
			err: nil,
		},
		{
			test:    "Cannot list roles",
			dirName: "iam_role_policy_empty",
			mocks: func(client *awstest.MockFakeIAM) {
				client.On("ListRolesPages",
					&iam.ListRolesInput{},
					mock.MatchedBy(func(callback func(res *iam.ListRolesOutput, lastPage bool) bool) bool {
						return true
					})).Return(awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamRolePolicyResourceType, resourceaws.AwsIamRoleResourceType),
		},
	}
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
		resourceaws.InitResourcesMetadata(repo)
		factory := terraform.NewTerraformResourceFactory(repo)

		deserializer := resource.NewDeserializer(factory)
		if shouldUpdate {
			provider, err := InitTestAwsProvider(providerLibrary)
			if err != nil {
				t.Fatal(err)
			}
			supplierLibrary.AddSupplier(NewIamRolePolicySupplier(provider, deserializer))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeIam := awstest.MockFakeIAM{}
			c.mocks(&fakeIam)

			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
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

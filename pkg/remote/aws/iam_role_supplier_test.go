package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/test/mocks"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
)

func TestIamRoleSupplier_Resources(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockIAMRepository)
		err     error
	}{
		{
			test:    "no iam roles",
			dirName: "iam_role_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Return([]*iam.Role{}, nil)
			},
			err: nil,
		},
		{
			test:    "iam multiples roles",
			dirName: "iam_role_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Return([]*iam.Role{
					{
						RoleName: aws.String("test_role_0"),
					},
					{
						RoleName: aws.String("test_role_1"),
					},
					{
						RoleName: aws.String("test_role_2"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "iam roles ignore services roles",
			dirName: "iam_role_ignore_services_roles",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Return([]*iam.Role{
					{
						RoleName: aws.String("AWSServiceRoleForOrganizations"),
					},
					{
						RoleName: aws.String("AWSServiceRoleForSupport"),
					},
					{
						RoleName: aws.String("AWSServiceRoleForTrustedAdvisor"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list iam roles",
			dirName: "iam_role_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllRoles").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamRoleResourceType),
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
			supplierLibrary.AddSupplier(NewIamRoleSupplier(provider, deserializer, repository.NewIAMRepository(provider.session)))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeIam := repository.MockIAMRepository{}
			c.mocks(&fakeIam)

			provider := mocks.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			s := &IamRoleSupplier{
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

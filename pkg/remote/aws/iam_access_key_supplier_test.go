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

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/test/mocks"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
)

func TestIamAccessKeySupplier_Resources(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockIAMRepository)
		err     error
	}{
		{
			test:    "no iam access_key",
			dirName: "iam_access_key_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				users := []*iam.User{
					{
						UserName: aws.String("test-driftctl"),
					},
				}
				repo.On("ListAllUsers").Return(users, nil)
				repo.On("ListAllAccessKeys", users).Return([]*iam.AccessKeyMetadata{}, nil)
			},
			err: nil,
		},
		{
			test:    "iam multiples keys for multiples users",
			dirName: "iam_access_key_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				users := []*iam.User{
					{
						UserName: aws.String("test-driftctl"),
					},
				}
				repo.On("ListAllUsers").Return(users, nil)
				repo.On("ListAllAccessKeys", users).Return([]*iam.AccessKeyMetadata{
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
				}, nil)
			},
			err: nil,
		},
		{
			test:    "Cannot list iam user",
			dirName: "iam_access_key_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllUsers").Once().Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamAccessKeyResourceType, resourceaws.AwsIamUserResourceType),
		},
		{
			test:    "Cannot list iam access_key",
			dirName: "iam_access_key_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllUsers").Once().Return([]*iam.User{}, nil)
				repo.On("ListAllAccessKeys", mock.Anything).Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamAccessKeyResourceType),
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
			supplierLibrary.AddSupplier(NewIamAccessKeySupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeIam := repository.MockIAMRepository{}
			c.mocks(&fakeIam)

			provider := mocks.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewIamAccessKeyDeserializer()
			s := &IamAccessKeySupplier{
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

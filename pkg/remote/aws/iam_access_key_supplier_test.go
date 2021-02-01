package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/iam"

	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/mocks"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
)

func TestIamAccessKeySupplier_Resources(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(client *mocks.FakeIAM)
		err     error
	}{
		{
			test:    "no iam access_key",
			dirName: "iam_access_key_empty",
			mocks: func(client *mocks.FakeIAM) {
				client.On("ListUsersPages",
					&iam.ListUsersInput{},
					mock.MatchedBy(func(callback func(res *iam.ListUsersOutput, lastPage bool) bool) bool {
						callback(&iam.ListUsersOutput{Users: []*iam.User{
							{
								UserName: aws.String("test-driftctl"),
							},
						}}, true)
						return true
					})).Return(nil)
				client.On("ListAccessKeysPages", mock.Anything, mock.Anything).Return(nil)
			},
			err: nil,
		},
		{
			test:    "iam multiples keys for multiples users",
			dirName: "iam_access_key_multiple",
			mocks: func(client *mocks.FakeIAM) {
				client.On("ListUsersPages",
					&iam.ListUsersInput{},
					mock.MatchedBy(func(callback func(res *iam.ListUsersOutput, lastPage bool) bool) bool {
						callback(&iam.ListUsersOutput{Users: []*iam.User{
							{
								UserName: aws.String("test-driftctl"),
							},
							{
								UserName: aws.String("test-driftctl2"),
							},
						}}, true)
						return true
					})).Return(nil)
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
			err: nil,
		},
		{
			test:    "Cannot list iam user",
			dirName: "iam_access_key_empty",
			mocks: func(client *mocks.FakeIAM) {
				client.On("ListUsersPages",
					&iam.ListUsersInput{},
					mock.MatchedBy(func(callback func(res *iam.ListUsersOutput, lastPage bool) bool) bool {
						return true
					})).Return(awserr.NewRequestFailure(nil, 403, ""))
				client.On("ListAccessKeysPages", mock.Anything, mock.Anything).Return(awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamAccessKeyResourceType, resourceaws.AwsIamUserResourceType),
		},
		{
			test:    "Cannot list iam access_key",
			dirName: "iam_access_key_empty",
			mocks: func(client *mocks.FakeIAM) {
				client.On("ListUsersPages",
					&iam.ListUsersInput{},
					mock.MatchedBy(func(callback func(res *iam.ListUsersOutput, lastPage bool) bool) bool {
						callback(&iam.ListUsersOutput{Users: []*iam.User{
							{
								UserName: aws.String("test-driftctl"),
							},
						}}, true)
						return true
					})).Return(nil)
				client.On("ListAccessKeysPages", mock.Anything, mock.Anything).Return(awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamAccessKeyResourceType),
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
			supplierLibrary.AddSupplier(NewIamAccessKeySupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeIam := mocks.FakeIAM{}
			c.mocks(&fakeIam)

			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
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

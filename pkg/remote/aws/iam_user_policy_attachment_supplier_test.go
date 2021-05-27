package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/cloudskiff/driftctl/test/mocks"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
)

func TestIamUserPolicyAttachmentSupplier_Resources(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockIAMRepository)
		err     error
	}{
		{
			test:    "no iam user policy",
			dirName: "iam_user_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				users := []*iam.User{
					{
						UserName: aws.String("loadbalancer"),
					},
				}
				repo.On("ListAllUsers").Return(users, nil)
				repo.On("ListAllUserPolicyAttachments", users).Return([]*repository.AttachedUserPolicy{}, nil)
			},
			err: nil,
		},
		{
			test:    "iam multiples users multiple policies",
			dirName: "iam_user_policy_attachment_multiple",
			mocks: func(repo *repository.MockIAMRepository) {
				users := []*iam.User{
					{
						UserName: aws.String("loadbalancer"),
					},
					{
						UserName: aws.String("loadbalancer2"),
					},
					{
						UserName: aws.String("loadbalancer3"),
					},
				}
				repo.On("ListAllUsers").Return(users, nil)
				repo.On("ListAllUserPolicyAttachments", users).Return([]*repository.AttachedUserPolicy{
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test"),
							PolicyName: aws.String("test-attach"),
						},
						UserName: *aws.String("loadbalancer"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test2"),
							PolicyName: aws.String("test-attach2"),
						},
						UserName: *aws.String("loadbalancer"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test3"),
							PolicyName: aws.String("test-attach3"),
						},
						UserName: *aws.String("loadbalancer"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test4"),
							PolicyName: aws.String("test-attach4"),
						},
						UserName: *aws.String("loadbalancer"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test"),
							PolicyName: aws.String("test-attach"),
						},
						UserName: *aws.String("loadbalancer2"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test2"),
							PolicyName: aws.String("test-attach2"),
						},
						UserName: *aws.String("loadbalancer2"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test3"),
							PolicyName: aws.String("test-attach3"),
						},
						UserName: *aws.String("loadbalancer2"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test4"),
							PolicyName: aws.String("test-attach4"),
						},
						UserName: *aws.String("loadbalancer2"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test"),
							PolicyName: aws.String("test-attach"),
						},
						UserName: *aws.String("loadbalancer3"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test2"),
							PolicyName: aws.String("test-attach2"),
						},
						UserName: *aws.String("loadbalancer3"),
					},
					{
						AttachedPolicy: iam.AttachedPolicy{
							PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test3"),
							PolicyName: aws.String("test-attach3"),
						},
						UserName: *aws.String("loadbalancer3"),
					},
				}, nil)

			},
			err: nil,
		},
		{
			test:    "cannot list user",
			dirName: "iam_user_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllUsers").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamUserPolicyAttachmentResourceType, resourceaws.AwsIamUserResourceType),
		},
		{
			test:    "cannot list user policies attachment",
			dirName: "iam_user_policy_empty",
			mocks: func(repo *repository.MockIAMRepository) {
				repo.On("ListAllUsers").Once().Return([]*iam.User{}, nil)
				repo.On("ListAllUserPolicyAttachments", mock.Anything).Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamUserPolicyAttachmentResourceType),
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
			supplierLibrary.AddSupplier(NewIamUserPolicyAttachmentSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeIam := repository.MockIAMRepository{}
			c.mocks(&fakeIam)

			provider := mocks.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewIamUserPolicyAttachmentDeserializer()
			s := &IamUserPolicyAttachmentSupplier{
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

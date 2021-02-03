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

func TestIamUserPolicyAttachmentSupplier_Resources(t *testing.T) {

	cases := []struct {
		test    string
		dirName string
		mocks   func(client *mocks.FakeIAM)
		err     error
	}{
		{
			test:    "iam multiples users multiple policies",
			dirName: "iam_user_policy_attachment_multiple",
			mocks: func(client *mocks.FakeIAM) {
				client.On("ListUsersPages",
					&iam.ListUsersInput{},
					mock.MatchedBy(func(callback func(res *iam.ListUsersOutput, lastPage bool) bool) bool {
						callback(&iam.ListUsersOutput{Users: []*iam.User{
							{
								UserName: aws.String("loadbalancer"),
							},
							{
								UserName: aws.String("loadbalancer2"),
							},
							{
								UserName: aws.String("loadbalancer3"),
							},
						}}, true)
						return true
					})).Return(nil).Once()

				shouldSkipfirst := false
				shouldSkipSecond := false
				shouldSkipThird := false

				client.On("ListAttachedUserPoliciesPages",
					&iam.ListAttachedUserPoliciesInput{
						UserName: aws.String("loadbalancer"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListAttachedUserPoliciesOutput, lastPage bool) bool) bool {
						if shouldSkipfirst {
							return false
						}
						callback(&iam.ListAttachedUserPoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test"),
								PolicyName: aws.String("test-attach"),
							},
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test2"),
								PolicyName: aws.String("test-attach2"),
							},
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test3"),
								PolicyName: aws.String("test-attach3"),
							},
						}}, false)
						callback(&iam.ListAttachedUserPoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test4"),
								PolicyName: aws.String("test-attach4"),
							},
						}}, true)
						shouldSkipfirst = true
						return true
					})).Return(nil).Once()

				client.On("ListAttachedUserPoliciesPages",
					&iam.ListAttachedUserPoliciesInput{
						UserName: aws.String("loadbalancer2"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListAttachedUserPoliciesOutput, lastPage bool) bool) bool {
						if shouldSkipSecond {
							return false
						}
						callback(&iam.ListAttachedUserPoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test"),
								PolicyName: aws.String("test-attach"),
							},
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test2"),
								PolicyName: aws.String("test-attach2"),
							},
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test3"),
								PolicyName: aws.String("test-attach3"),
							},
						}}, false)
						callback(&iam.ListAttachedUserPoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test4"),
								PolicyName: aws.String("test-attach4"),
							},
						}}, true)
						shouldSkipSecond = true
						return true
					})).Return(nil).Once()

				client.On("ListAttachedUserPoliciesPages",
					&iam.ListAttachedUserPoliciesInput{
						UserName: aws.String("loadbalancer3"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListAttachedUserPoliciesOutput, lastPage bool) bool) bool {
						if shouldSkipThird {
							return false
						}
						callback(&iam.ListAttachedUserPoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test"),
								PolicyName: aws.String("test-attach"),
							},
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test2"),
								PolicyName: aws.String("test-attach2"),
							},
							&iam.AttachedPolicy{
								PolicyArn:  aws.String("arn:aws:iam::526954929923:policy/test3"),
								PolicyName: aws.String("test-attach3"),
							},
						}}, false)
						shouldSkipThird = true
						return true
					})).Return(nil).Once()

			},
			err: nil,
		},
		{
			test:    "cannot list user",
			dirName: "iam_user_policy_empty",
			mocks: func(client *mocks.FakeIAM) {
				client.On("ListUsersPages",
					&iam.ListUsersInput{},
					mock.MatchedBy(func(callback func(res *iam.ListUsersOutput, lastPage bool) bool) bool {
						return true
					})).Return(awserr.NewRequestFailure(nil, 403, "")).Once()
			},
			err: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamUserPolicyAttachmentResourceType, resourceaws.AwsIamUserResourceType),
		},
		{
			test:    "cannot list user policies attachment",
			dirName: "iam_user_policy_empty",
			mocks: func(client *mocks.FakeIAM) {
				client.On("ListUsersPages",
					&iam.ListUsersInput{},
					mock.MatchedBy(func(callback func(res *iam.ListUsersOutput, lastPage bool) bool) bool {
						callback(&iam.ListUsersOutput{Users: []*iam.User{
							{
								UserName: aws.String("loadbalancer"),
							},
							{
								UserName: aws.String("loadbalancer2"),
							},
							{
								UserName: aws.String("loadbalancer3"),
							},
						}}, true)
						return true
					})).Return(nil).Once()
				client.On("ListAttachedUserPoliciesPages",
					mock.Anything,
					mock.MatchedBy(func(callback func(res *iam.ListAttachedUserPoliciesOutput, lastPage bool) bool) bool {
						return true
					})).Return(awserr.NewRequestFailure(nil, 403, "")).Once()
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsIamUserPolicyAttachmentResourceType),
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
			supplierLibrary.AddSupplier(NewIamUserPolicyAttachmentSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeIam := mocks.FakeIAM{}
			c.mocks(&fakeIam)

			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
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

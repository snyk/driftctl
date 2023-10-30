package remote

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/enumeration/terraform"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/mocks"

	"github.com/snyk/driftctl/test/goldenfile"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCloudformationStack(t *testing.T) {
	awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 400, "")
	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockCloudformationRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			test:    "no cloudformation stacks",
			dirName: "aws_cloudformation_stack_empty",
			mocks: func(repository *repository.MockCloudformationRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllStacks").Return([]*cloudformation.Stack{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test:    "multiple cloudformation stacks",
			dirName: "aws_cloudformation_stack_multiple",
			mocks: func(repository *repository.MockCloudformationRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllStacks").Return([]*cloudformation.Stack{
					{StackId: awssdk.String("arn:aws:cloudformation:us-east-1:047081014315:stack/bar-stack/c7a96e70-0f21-11ec-bd2a-0a2d95c2b2ab")},
					{StackId: awssdk.String("arn:aws:cloudformation:us-east-1:047081014315:stack/foo-stack/c7aa0ab0-0f21-11ec-ba25-129d8c0b3757")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "arn:aws:cloudformation:us-east-1:047081014315:stack/bar-stack/c7a96e70-0f21-11ec-bd2a-0a2d95c2b2ab", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsCloudformationStackResourceType, got[0].ResourceType())

				assert.Equal(t, "arn:aws:cloudformation:us-east-1:047081014315:stack/foo-stack/c7aa0ab0-0f21-11ec-ba25-129d8c0b3757", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsCloudformationStackResourceType, got[1].ResourceType())
			},
		},
		{
			test:    "cannot list cloudformation stacks",
			dirName: "aws_cloudformation_stack_list",
			mocks: func(repository *repository.MockCloudformationRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllStacks").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsCloudformationStackResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsCloudformationStackResourceType, resourceaws.AwsCloudformationStackResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockCloudformationRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.CloudformationRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewCloudformationRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewCloudformationStackEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

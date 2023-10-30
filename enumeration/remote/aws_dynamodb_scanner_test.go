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
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/mocks"

	"github.com/snyk/driftctl/test/goldenfile"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDynamoDBTable(t *testing.T) {
	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockDynamoDBRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			test:    "no DynamoDB Table",
			dirName: "aws_dynamodb_table_empty",
			mocks: func(client *repository.MockDynamoDBRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllTables").Return([]*string{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
		{
			test:    "Multiple DynamoDB Table",
			dirName: "aws_dynamodb_table_multiple",
			mocks: func(client *repository.MockDynamoDBRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllTables").Return([]*string{
					awssdk.String("GameScores"),
					awssdk.String("example"),
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "GameScores", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsDynamodbTableResourceType, got[0].ResourceType())

				assert.Equal(t, "example", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsDynamodbTableResourceType, got[1].ResourceType())
			},
			wantErr: nil,
		},
		{
			test:    "cannot list DynamoDB Table",
			dirName: "aws_dynamodb_table_list",
			mocks: func(client *repository.MockDynamoDBRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 400, "")
				client.On("ListAllTables").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsDynamodbTableResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsDynamodbTableResourceType, resourceaws.AwsDynamodbTableResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
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
			fakeRepo := &repository.MockDynamoDBRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.DynamoDBRepository = fakeRepo
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
				repo = repository.NewDynamoDBRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewDynamoDBTableEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

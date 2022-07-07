package remote

import (
	"errors"
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/mocks"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/goldenfile"
	testresource "github.com/snyk/driftctl/test/resource"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDynamoDBTable(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockDynamoDBRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no DynamoDB Table",
			dirName: "aws_dynamodb_table_empty",
			mocks: func(client *repository.MockDynamoDBRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllTables").Return([]*string{}, nil)
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
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
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
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsDynamodbTableResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsDynamodbTableResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsDynamodbTableResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

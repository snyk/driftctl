package remote

import (
	"errors"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDynamoDBTable(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockDynamoDBRepository)
		wantErr error
	}{
		{
			test:    "no DynamoDB Table",
			dirName: "dynamodb_table_empty",
			mocks: func(client *repository.MockDynamoDBRepository) {
				client.On("ListAllTables").Return([]*string{}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "Multiple DynamoDB Table",
			dirName: "dynamodb_table_multiple",
			mocks: func(client *repository.MockDynamoDBRepository) {
				client.On("ListAllTables").Return([]*string{
					awssdk.String("GameScores"),
					awssdk.String("example"),
				}, nil)
			},
			wantErr: nil,
		},
		{
			test:    "cannot list DynamoDB Table",
			dirName: "dynamodb_table_list",
			mocks: func(client *repository.MockDynamoDBRepository) {
				client.On("ListAllTables").Return(nil, awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 400, ""))
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
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
			fakeRepo := &repository.MockDynamoDBRepository{}
			c.mocks(fakeRepo)
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
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsDynamodbTableResourceType, aws.NewDynamoDBTableDetailsFetcher(provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsDynamodbTableResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

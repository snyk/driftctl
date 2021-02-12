package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/mocks"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
)

func TestDynamoDBTableSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *mocks.DynamoDBRepository)
		err     error
	}{
		{
			test:    "no DynamoDB Table",
			dirName: "dynamodb_table_empty",
			mocks: func(client *mocks.DynamoDBRepository) {
				client.On("ListAllTables").Return([]*string{}, nil)
			},
			err: nil,
		},
		{
			test:    "Multiple DynamoDB Table",
			dirName: "dynamodb_table_multiple",
			mocks: func(client *mocks.DynamoDBRepository) {
				client.On("ListAllTables").Return([]*string{
					aws.String("GameScores"),
					aws.String("example"),
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list DynamoDB Table",
			dirName: "dynamodb_table_list",
			mocks: func(client *mocks.DynamoDBRepository) {
				client.On("ListAllTables").Return(nil, awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 400, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 400, ""), resourceaws.AwsDynamodbTableResourceType),
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

			supplierLibrary.AddSupplier(NewDynamoDBTableSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeClient := mocks.DynamoDBRepository{}
			c.mocks(&fakeClient)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			dynamoDBTableDeserializer := awsdeserializer.NewDynamoDBTableDeserializer()
			s := &DynamoDBTableSupplier{
				provider,
				dynamoDBTableDeserializer,
				&fakeClient,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, dynamoDBTableDeserializer, shouldUpdate, tt)
		})
	}
}

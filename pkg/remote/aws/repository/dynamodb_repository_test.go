package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/snyk/driftctl/pkg/remote/cache"
	awstest "github.com/snyk/driftctl/test/aws"

	"github.com/stretchr/testify/mock"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_dynamoDBRepository_ListAllTopics(t *testing.T) {

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeDynamoDB)
		want    []*string
		wantErr error
	}{
		{
			name: "List with 2 pages",
			mocks: func(client *awstest.MockFakeDynamoDB) {
				client.On("ListTablesPages",
					&dynamodb.ListTablesInput{},
					mock.MatchedBy(func(callback func(res *dynamodb.ListTablesOutput, lastPage bool) bool) bool {
						callback(&dynamodb.ListTablesOutput{
							TableNames: []*string{
								aws.String("1"),
								aws.String("2"),
								aws.String("3"),
							},
						}, false)
						callback(&dynamodb.ListTablesOutput{
							TableNames: []*string{
								aws.String("4"),
								aws.String("5"),
								aws.String("6"),
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*string{
				aws.String("1"),
				aws.String("2"),
				aws.String("3"),
				aws.String("4"),
				aws.String("5"),
				aws.String("6"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := awstest.MockFakeDynamoDB{}
			tt.mocks(&client)
			r := &dynamoDBRepository{
				client: &client,
				cache:  store,
			}
			got, err := r.ListAllTables()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllTables()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*string{}, store.Get("dynamodbListAllTables"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

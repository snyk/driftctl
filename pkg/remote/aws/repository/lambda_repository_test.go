package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	awstest "github.com/cloudskiff/driftctl/test/aws"
	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_lambdaRepository_ListAllLambdaFunctions(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeLambda)
		want    []*lambda.FunctionConfiguration
		wantErr error
	}{
		{
			name: "List with 2 pages",
			mocks: func(client *awstest.MockFakeLambda) {
				client.On("ListFunctionsPages",
					&lambda.ListFunctionsInput{},
					mock.MatchedBy(func(callback func(res *lambda.ListFunctionsOutput, lastPage bool) bool) bool {
						callback(&lambda.ListFunctionsOutput{
							Functions: []*lambda.FunctionConfiguration{
								{FunctionName: aws.String("1")},
								{FunctionName: aws.String("2")},
								{FunctionName: aws.String("3")},
								{FunctionName: aws.String("4")},
							},
						}, false)
						callback(&lambda.ListFunctionsOutput{
							Functions: []*lambda.FunctionConfiguration{
								{FunctionName: aws.String("5")},
								{FunctionName: aws.String("6")},
								{FunctionName: aws.String("7")},
								{FunctionName: aws.String("8")},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*lambda.FunctionConfiguration{
				{FunctionName: aws.String("1")},
				{FunctionName: aws.String("2")},
				{FunctionName: aws.String("3")},
				{FunctionName: aws.String("4")},
				{FunctionName: aws.String("5")},
				{FunctionName: aws.String("6")},
				{FunctionName: aws.String("7")},
				{FunctionName: aws.String("8")},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeLambda{}
			tt.mocks(client)
			r := &lambdaRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllLambdaFunctions()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllLambdaFunctions()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*lambda.FunctionConfiguration{}, store.Get("lambdaListAllLambdaFunctions"))
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

func Test_lambdaRepository_ListAllLambdaEventSourceMappings(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(mock *awstest.MockFakeLambda)
		want    []*lambda.EventSourceMappingConfiguration
		wantErr error
	}{
		{
			name: "List with 2 pages",
			mocks: func(client *awstest.MockFakeLambda) {
				client.On("ListEventSourceMappingsPages",
					&lambda.ListEventSourceMappingsInput{},
					mock.MatchedBy(func(callback func(res *lambda.ListEventSourceMappingsOutput, lastPage bool) bool) bool {
						callback(&lambda.ListEventSourceMappingsOutput{
							EventSourceMappings: []*lambda.EventSourceMappingConfiguration{
								{UUID: aws.String("1")},
								{UUID: aws.String("2")},
								{UUID: aws.String("3")},
								{UUID: aws.String("4")},
							},
						}, false)
						callback(&lambda.ListEventSourceMappingsOutput{
							EventSourceMappings: []*lambda.EventSourceMappingConfiguration{
								{UUID: aws.String("5")},
								{UUID: aws.String("6")},
								{UUID: aws.String("7")},
								{UUID: aws.String("8")},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*lambda.EventSourceMappingConfiguration{
				{UUID: aws.String("1")},
				{UUID: aws.String("2")},
				{UUID: aws.String("3")},
				{UUID: aws.String("4")},
				{UUID: aws.String("5")},
				{UUID: aws.String("6")},
				{UUID: aws.String("7")},
				{UUID: aws.String("8")},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeLambda{}
			tt.mocks(client)
			r := &lambdaRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllLambdaEventSourceMappings()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllLambdaEventSourceMappings()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*lambda.EventSourceMappingConfiguration{}, store.Get("lambdaListAllLambdaEventSourceMappings"))
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

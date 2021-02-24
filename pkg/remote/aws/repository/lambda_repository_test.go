package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_lambdaRepository_ListAllLambdaFunctions(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(mock *MockLambdaClient)
		want    []*lambda.FunctionConfiguration
		wantErr error
	}{
		{
			name: "List with 2 pages",
			mocks: func(client *MockLambdaClient) {
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
					})).Return(nil)
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
			client := &MockLambdaClient{}
			tt.mocks(client)
			r := &lambdaRepository{
				client: client,
			}
			got, err := r.ListAllLambdaFunctions()
			assert.Equal(t, tt.wantErr, err)
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

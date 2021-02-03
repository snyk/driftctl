package mocks

import (
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
)

type ListFunctionsPagesOutput []struct {
	LastPage bool
	Response *lambda.ListFunctionsOutput
}

type MockAWSLambdaClient struct {
	lambdaiface.LambdaAPI
	functionsPages ListFunctionsPagesOutput
	err            error
}

func NewMockAWSLambdaErrorClient(err error) *MockAWSLambdaClient {
	return &MockAWSLambdaClient{err: err}
}

func NewMockAWSLambdaClient(functionsPages ListFunctionsPagesOutput) *MockAWSLambdaClient {
	return &MockAWSLambdaClient{functionsPages: functionsPages}
}

func (m *MockAWSLambdaClient) ListFunctionsPages(_ *lambda.ListFunctionsInput, cb func(*lambda.ListFunctionsOutput, bool) bool) error {
	if m.err != nil {
		return m.err
	}
	for _, functionsPage := range m.functionsPages {
		cb(functionsPage.Response, functionsPage.LastPage)
	}
	return nil
}

package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/snyk/driftctl/pkg/remote/cache"
)

type LambdaRepository interface {
	ListAllLambdaFunctions() ([]*lambda.FunctionConfiguration, error)
	ListAllLambdaEventSourceMappings() ([]*lambda.EventSourceMappingConfiguration, error)
}

type lambdaRepository struct {
	client lambdaiface.LambdaAPI
	cache  cache.Cache
}

func NewLambdaRepository(session *session.Session, c cache.Cache) *lambdaRepository {
	return &lambdaRepository{
		lambda.New(session),
		c,
	}
}

func (r *lambdaRepository) ListAllLambdaFunctions() ([]*lambda.FunctionConfiguration, error) {
	if v := r.cache.Get("lambdaListAllLambdaFunctions"); v != nil {
		return v.([]*lambda.FunctionConfiguration), nil
	}

	var functions []*lambda.FunctionConfiguration
	input := &lambda.ListFunctionsInput{}
	err := r.client.ListFunctionsPages(input, func(res *lambda.ListFunctionsOutput, lastPage bool) bool {
		functions = append(functions, res.Functions...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put("lambdaListAllLambdaFunctions", functions)
	return functions, nil
}

func (r *lambdaRepository) ListAllLambdaEventSourceMappings() ([]*lambda.EventSourceMappingConfiguration, error) {
	if v := r.cache.Get("lambdaListAllLambdaEventSourceMappings"); v != nil {
		return v.([]*lambda.EventSourceMappingConfiguration), nil
	}

	var eventSourceMappingConfigurations []*lambda.EventSourceMappingConfiguration
	input := &lambda.ListEventSourceMappingsInput{}
	err := r.client.ListEventSourceMappingsPages(input, func(res *lambda.ListEventSourceMappingsOutput, lastPage bool) bool {
		eventSourceMappingConfigurations = append(eventSourceMappingConfigurations, res.EventSourceMappings...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put("lambdaListAllLambdaEventSourceMappings", eventSourceMappingConfigurations)
	return eventSourceMappingConfigurations, nil
}

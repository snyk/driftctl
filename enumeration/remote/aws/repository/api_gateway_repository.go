package repository

import (
	"fmt"
	"github.com/snyk/driftctl/enumeration/remote/cache"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/apigateway/apigatewayiface"
)

type ApiGatewayRepository interface {
	ListAllRestApis() ([]*apigateway.RestApi, error)
	GetAccount() (*apigateway.Account, error)
	ListAllApiKeys() ([]*apigateway.ApiKey, error)
	ListAllRestApiAuthorizers(string) ([]*apigateway.Authorizer, error)
	ListAllRestApiStages(string) ([]*apigateway.Stage, error)
	ListAllRestApiResources(string) ([]*apigateway.Resource, error)
	ListAllDomainNames() ([]*apigateway.DomainName, error)
	ListAllVpcLinks() ([]*apigateway.UpdateVpcLinkOutput, error)
	ListAllRestApiRequestValidators(string) ([]*apigateway.UpdateRequestValidatorOutput, error)
	ListAllDomainNameBasePathMappings(string) ([]*apigateway.BasePathMapping, error)
	ListAllRestApiModels(string) ([]*apigateway.Model, error)
	ListAllRestApiGatewayResponses(string) ([]*apigateway.UpdateGatewayResponseOutput, error)
}

type apigatewayRepository struct {
	client apigatewayiface.APIGatewayAPI
	cache  cache.Cache
}

func NewApiGatewayRepository(session *session.Session, c cache.Cache) *apigatewayRepository {
	return &apigatewayRepository{
		apigateway.New(session),
		c,
	}
}

func (r *apigatewayRepository) ListAllRestApis() ([]*apigateway.RestApi, error) {
	cacheKey := "apigatewayListAllRestApis"
	v := r.cache.GetAndLock(cacheKey)
	defer r.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*apigateway.RestApi), nil
	}

	var restApis []*apigateway.RestApi
	input := apigateway.GetRestApisInput{}
	err := r.client.GetRestApisPages(&input,
		func(resp *apigateway.GetRestApisOutput, lastPage bool) bool {
			restApis = append(restApis, resp.Items...)
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, restApis)
	return restApis, nil
}

func (r *apigatewayRepository) GetAccount() (*apigateway.Account, error) {
	if v := r.cache.Get("apigatewayGetAccount"); v != nil {
		return v.(*apigateway.Account), nil
	}

	account, err := r.client.GetAccount(&apigateway.GetAccountInput{})
	if err != nil {
		return nil, err
	}

	r.cache.Put("apigatewayGetAccount", account)
	return account, nil
}

func (r *apigatewayRepository) ListAllApiKeys() ([]*apigateway.ApiKey, error) {
	if v := r.cache.Get("apigatewayListAllApiKeys"); v != nil {
		return v.([]*apigateway.ApiKey), nil
	}

	var apiKeys []*apigateway.ApiKey
	input := apigateway.GetApiKeysInput{}
	err := r.client.GetApiKeysPages(&input,
		func(resp *apigateway.GetApiKeysOutput, lastPage bool) bool {
			apiKeys = append(apiKeys, resp.Items...)
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}

	r.cache.Put("apigatewayListAllApiKeys", apiKeys)
	return apiKeys, nil
}

func (r *apigatewayRepository) ListAllRestApiAuthorizers(apiId string) ([]*apigateway.Authorizer, error) {
	cacheKey := fmt.Sprintf("apigatewayListAllRestApiAuthorizers_api_%s", apiId)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*apigateway.Authorizer), nil
	}

	input := &apigateway.GetAuthorizersInput{
		RestApiId: &apiId,
	}
	resources, err := r.client.GetAuthorizers(input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources.Items)
	return resources.Items, nil
}

func (r *apigatewayRepository) ListAllRestApiStages(apiId string) ([]*apigateway.Stage, error) {
	cacheKey := fmt.Sprintf("apigatewayListAllRestApiStages_api_%s", apiId)
	v := r.cache.GetAndLock(cacheKey)
	defer r.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*apigateway.Stage), nil
	}

	input := &apigateway.GetStagesInput{
		RestApiId: &apiId,
	}
	resources, err := r.client.GetStages(input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources.Item)
	return resources.Item, nil
}

func (r *apigatewayRepository) ListAllRestApiResources(apiId string) ([]*apigateway.Resource, error) {
	cacheKey := fmt.Sprintf("apigatewayListAllRestApiResources_api_%s", apiId)
	v := r.cache.GetAndLock(cacheKey)
	defer r.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*apigateway.Resource), nil
	}

	var resources []*apigateway.Resource
	input := &apigateway.GetResourcesInput{
		RestApiId: &apiId,
		Embed:     []*string{aws.String("methods")},
	}
	err := r.client.GetResourcesPages(input, func(res *apigateway.GetResourcesOutput, lastPage bool) bool {
		resources = append(resources, res.Items...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources)
	return resources, nil
}

func (r *apigatewayRepository) ListAllDomainNames() ([]*apigateway.DomainName, error) {
	cacheKey := "apigatewayListAllDomainNames"
	v := r.cache.GetAndLock(cacheKey)
	defer r.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*apigateway.DomainName), nil
	}

	var domainNames []*apigateway.DomainName
	input := apigateway.GetDomainNamesInput{}
	err := r.client.GetDomainNamesPages(&input,
		func(resp *apigateway.GetDomainNamesOutput, lastPage bool) bool {
			domainNames = append(domainNames, resp.Items...)
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, domainNames)
	return domainNames, nil
}

func (r *apigatewayRepository) ListAllVpcLinks() ([]*apigateway.UpdateVpcLinkOutput, error) {
	if v := r.cache.Get("apigatewayListAllVpcLinks"); v != nil {
		return v.([]*apigateway.UpdateVpcLinkOutput), nil
	}

	var vpcLinks []*apigateway.UpdateVpcLinkOutput
	input := apigateway.GetVpcLinksInput{}
	err := r.client.GetVpcLinksPages(&input,
		func(resp *apigateway.GetVpcLinksOutput, lastPage bool) bool {
			vpcLinks = append(vpcLinks, resp.Items...)
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}

	r.cache.Put("apigatewayListAllVpcLinks", vpcLinks)
	return vpcLinks, nil
}

func (r *apigatewayRepository) ListAllRestApiRequestValidators(apiId string) ([]*apigateway.UpdateRequestValidatorOutput, error) {
	cacheKey := fmt.Sprintf("apigatewayListAllRestApiRequestValidators_api_%s", apiId)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*apigateway.UpdateRequestValidatorOutput), nil
	}

	input := &apigateway.GetRequestValidatorsInput{
		RestApiId: &apiId,
	}
	resources, err := r.client.GetRequestValidators(input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources.Items)
	return resources.Items, nil
}

func (r *apigatewayRepository) ListAllDomainNameBasePathMappings(domainName string) ([]*apigateway.BasePathMapping, error) {
	cacheKey := fmt.Sprintf("apigatewayListAllDomainNameBasePathMappings_domainName_%s", domainName)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*apigateway.BasePathMapping), nil
	}

	var mappings []*apigateway.BasePathMapping
	input := &apigateway.GetBasePathMappingsInput{
		DomainName: &domainName,
	}
	err := r.client.GetBasePathMappingsPages(input, func(res *apigateway.GetBasePathMappingsOutput, lastPage bool) bool {
		mappings = append(mappings, res.Items...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, mappings)
	return mappings, nil
}

func (r *apigatewayRepository) ListAllRestApiModels(apiId string) ([]*apigateway.Model, error) {
	cacheKey := fmt.Sprintf("apigatewayListAllRestApiModels_api_%s", apiId)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*apigateway.Model), nil
	}

	var resources []*apigateway.Model
	input := &apigateway.GetModelsInput{
		RestApiId: &apiId,
	}
	err := r.client.GetModelsPages(input, func(res *apigateway.GetModelsOutput, lastPage bool) bool {
		resources = append(resources, res.Items...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources)
	return resources, nil
}

func (r *apigatewayRepository) ListAllRestApiGatewayResponses(apiId string) ([]*apigateway.UpdateGatewayResponseOutput, error) {
	cacheKey := fmt.Sprintf("apigatewayListAllRestApiGatewayResponses_api_%s", apiId)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*apigateway.UpdateGatewayResponseOutput), nil
	}

	input := &apigateway.GetGatewayResponsesInput{
		RestApiId: &apiId,
	}
	resources, err := r.client.GetGatewayResponses(input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources.Items)
	return resources.Items, nil
}

package repository

import (
	"fmt"
	"github.com/snyk/driftctl/enumeration/remote/cache"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/aws/aws-sdk-go/service/apigatewayv2/apigatewayv2iface"
)

type ApiGatewayV2Repository interface {
	ListAllApis() ([]*apigatewayv2.Api, error)
	ListAllApiRoutes(apiId *string) ([]*apigatewayv2.Route, error)
	ListAllApiDeployments(apiId *string) ([]*apigatewayv2.Deployment, error)
	ListAllVpcLinks() ([]*apigatewayv2.VpcLink, error)
	ListAllApiAuthorizers(string) ([]*apigatewayv2.Authorizer, error)
	ListAllApiIntegrations(string) ([]*apigatewayv2.Integration, error)
	ListAllApiModels(string) ([]*apigatewayv2.Model, error)
	ListAllApiStages(string) ([]*apigatewayv2.Stage, error)
	ListAllApiRouteResponses(string, string) ([]*apigatewayv2.RouteResponse, error)
	ListAllApiMappings(string) ([]*apigatewayv2.ApiMapping, error)
	ListAllApiIntegrationResponses(string, string) ([]*apigatewayv2.IntegrationResponse, error)
}
type apigatewayv2Repository struct {
	client apigatewayv2iface.ApiGatewayV2API
	cache  cache.Cache
}

func NewApiGatewayV2Repository(session *session.Session, c cache.Cache) *apigatewayv2Repository {
	return &apigatewayv2Repository{
		apigatewayv2.New(session),
		c,
	}
}

func (r *apigatewayv2Repository) ListAllApis() ([]*apigatewayv2.Api, error) {
	cacheKey := "apigatewayv2ListAllApis"
	v := r.cache.GetAndLock(cacheKey)
	defer r.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*apigatewayv2.Api), nil
	}

	input := apigatewayv2.GetApisInput{}
	resources, err := r.client.GetApis(&input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources.Items)
	return resources.Items, nil
}

func (r *apigatewayv2Repository) ListAllApiRoutes(apiID *string) ([]*apigatewayv2.Route, error) {
	cacheKey := fmt.Sprintf("apigatewayv2ListAllApiRoutes_api_%s", *apiID)
	v := r.cache.GetAndLock(cacheKey)
	defer r.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*apigatewayv2.Route), nil
	}

	resources, err := r.client.GetRoutes(&apigatewayv2.GetRoutesInput{ApiId: apiID})
	if err != nil {
		return nil, err
	}
	r.cache.Put(cacheKey, resources.Items)
	return resources.Items, nil
}

func (r *apigatewayv2Repository) ListAllApiDeployments(apiID *string) ([]*apigatewayv2.Deployment, error) {
	cacheKey := fmt.Sprintf("apigatewayv2ListAllApiDeployments_api_%s", *apiID)
	v := r.cache.Get(cacheKey)

	if v != nil {
		return v.([]*apigatewayv2.Deployment), nil
	}

	resources, err := r.client.GetDeployments(&apigatewayv2.GetDeploymentsInput{ApiId: apiID})
	if err != nil {
		return nil, err
	}
	r.cache.Put(cacheKey, resources.Items)
	return resources.Items, nil
}

func (r *apigatewayv2Repository) ListAllVpcLinks() ([]*apigatewayv2.VpcLink, error) {
	if v := r.cache.Get("apigatewayv2ListAllVpcLinks"); v != nil {
		return v.([]*apigatewayv2.VpcLink), nil
	}

	input := apigatewayv2.GetVpcLinksInput{}
	resources, err := r.client.GetVpcLinks(&input)
	if err != nil {
		return nil, err
	}

	r.cache.Put("apigatewayv2ListAllVpcLinks", resources.Items)
	return resources.Items, nil
}

func (r *apigatewayv2Repository) ListAllApiAuthorizers(apiId string) ([]*apigatewayv2.Authorizer, error) {
	cacheKey := fmt.Sprintf("apigatewayv2ListAllApiAuthorizers_api_%s", apiId)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*apigatewayv2.Authorizer), nil
	}

	input := apigatewayv2.GetAuthorizersInput{
		ApiId: &apiId,
	}
	resources, err := r.client.GetAuthorizers(&input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources.Items)
	return resources.Items, nil
}

func (r *apigatewayv2Repository) ListAllApiIntegrations(apiId string) ([]*apigatewayv2.Integration, error) {
	cacheKey := fmt.Sprintf("apigatewayv2ListAllApiIntegrations_api_%s", apiId)

	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*apigatewayv2.Integration), nil
	}

	input := apigatewayv2.GetIntegrationsInput{
		ApiId: &apiId,
	}
	resources, err := r.client.GetIntegrations(&input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources.Items)
	return resources.Items, nil
}

func (r *apigatewayv2Repository) ListAllApiModels(apiId string) ([]*apigatewayv2.Model, error) {
	cacheKey := fmt.Sprintf("apigatewayv2ListAllApiModels_api_%s", apiId)

	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*apigatewayv2.Model), nil
	}

	input := apigatewayv2.GetModelsInput{
		ApiId: &apiId,
	}
	resources, err := r.client.GetModels(&input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources.Items)
	return resources.Items, nil
}

func (r *apigatewayv2Repository) ListAllApiStages(apiId string) ([]*apigatewayv2.Stage, error) {
	cacheKey := fmt.Sprintf("apigatewayv2ListAllApiStages_api_%s", apiId)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*apigatewayv2.Stage), nil
	}

	input := apigatewayv2.GetStagesInput{
		ApiId: &apiId,
	}
	resources, err := r.client.GetStages(&input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources.Items)
	return resources.Items, nil
}

func (r *apigatewayv2Repository) ListAllApiIntegrationResponses(apiId, integrationId string) ([]*apigatewayv2.IntegrationResponse, error) {
	cacheKey := fmt.Sprintf("apigatewayv2ListAllApiIntegrationResponses_api_%s_integration_%s", apiId, integrationId)
	v := r.cache.Get(cacheKey)
	if v != nil {
		return v.([]*apigatewayv2.IntegrationResponse), nil
	}
	input := apigatewayv2.GetIntegrationResponsesInput{
		ApiId:         &apiId,
		IntegrationId: &integrationId,
	}
	resources, err := r.client.GetIntegrationResponses(&input)
	if err != nil {
		return nil, err
	}
	r.cache.Put(cacheKey, resources.Items)
	return resources.Items, nil
}

func (r *apigatewayv2Repository) ListAllApiRouteResponses(apiId, routeId string) ([]*apigatewayv2.RouteResponse, error) {
	cacheKey := fmt.Sprintf("apigatewayv2ListAllApiRouteResponses_api_%s_route_%s", apiId, routeId)
	v := r.cache.Get(cacheKey)
	if v != nil {
		return v.([]*apigatewayv2.RouteResponse), nil
	}
	input := apigatewayv2.GetRouteResponsesInput{
		ApiId:   &apiId,
		RouteId: &routeId,
	}
	resources, err := r.client.GetRouteResponses(&input)
	if err != nil {
		return nil, err
	}
	r.cache.Put(cacheKey, resources.Items)
	return resources.Items, nil
}

func (r *apigatewayv2Repository) ListAllApiMappings(domainName string) ([]*apigatewayv2.ApiMapping, error) {
	cacheKey := fmt.Sprintf("apigatewayv2ListAllApiMappings_api_%s", domainName)
	if v := r.cache.Get(cacheKey); v != nil {
		return v.([]*apigatewayv2.ApiMapping), nil
	}

	input := apigatewayv2.GetApiMappingsInput{
		DomainName: &domainName,
	}
	resources, err := r.client.GetApiMappings(&input)
	if err != nil {
		return nil, err
	}

	r.cache.Put(cacheKey, resources.Items)
	return resources.Items, nil
}

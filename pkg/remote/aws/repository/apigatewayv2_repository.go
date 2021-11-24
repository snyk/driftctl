package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/aws/aws-sdk-go/service/apigatewayv2/apigatewayv2iface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type ApiGatewayV2Repository interface {
	ListAllApis() ([]*apigatewayv2.Api, error)
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
	v := r.cache.Get(cacheKey)

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

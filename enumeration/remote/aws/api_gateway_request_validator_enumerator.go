package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type ApiGatewayRequestValidatorEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayRequestValidatorEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayRequestValidatorEnumerator {
	return &ApiGatewayRequestValidatorEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayRequestValidatorEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayRequestValidatorResourceType
}

func (e *ApiGatewayRequestValidatorEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllRestApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayRestApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, api := range apis {
		a := api
		requestValidators, err := e.repository.ListAllRestApiRequestValidators(*a.Id)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, requestValidator := range requestValidators {
			r := requestValidator
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*r.Id,
					map[string]interface{}{},
				),
			)
		}

	}
	return results, err
}

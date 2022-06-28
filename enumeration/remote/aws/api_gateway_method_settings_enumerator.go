package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"strings"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type ApiGatewayMethodSettingsEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayMethodSettingsEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayMethodSettingsEnumerator {
	return &ApiGatewayMethodSettingsEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayMethodSettingsEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayMethodSettingsResourceType
}

func (e *ApiGatewayMethodSettingsEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllRestApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayRestApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, api := range apis {
		a := api
		stages, err := e.repository.ListAllRestApiStages(*a.Id)
		if err != nil {
			return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayStageResourceType)
		}

		for _, stage := range stages {
			s := stage
			for methodPath := range s.MethodSettings {
				results = append(
					results,
					e.factory.CreateAbstractResource(
						string(e.SupportedType()),
						strings.Join([]string{*a.Id, *s.StageName, methodPath}, "-"),
						map[string]interface{}{},
					),
				)
			}
		}
	}

	return results, err
}

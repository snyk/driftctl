package aws

import (
	"strings"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type ApiGatewayStageEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayStageEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayStageEnumerator {
	return &ApiGatewayStageEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayStageEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayStageResourceType
}

func (e *ApiGatewayStageEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllRestApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayRestApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, api := range apis {
		a := api
		stages, err := e.repository.ListAllRestApiStages(*a.Id)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, stage := range stages {
			s := stage
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					strings.Join([]string{"ags", *a.Id, *s.StageName}, "-"),
					map[string]interface{}{},
				),
			)
		}

	}
	return results, err
}

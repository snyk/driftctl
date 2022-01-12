package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type ApiGatewayV2StageEnumerator struct {
	repository repository.ApiGatewayV2Repository
	factory    resource.ResourceFactory
}

func NewApiGatewayV2StageEnumerator(repo repository.ApiGatewayV2Repository, factory resource.ResourceFactory) *ApiGatewayV2StageEnumerator {
	return &ApiGatewayV2StageEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayV2StageEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayV2StageResourceType
}

func (e *ApiGatewayV2StageEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayV2ApiResourceType)
	}

	results := make([]*resource.Resource, 0)

	for _, api := range apis {
		stages, err := e.repository.ListAllApiStages(*api.ApiId)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}

		for _, stage := range stages {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*stage.StageName,
					map[string]interface{}{},
				),
			)
		}

	}

	return results, err
}

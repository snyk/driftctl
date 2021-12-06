package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type ApiGatewayRestApiEnumerator struct {
	repository repository.ApiGatewayRepository
	factory    resource.ResourceFactory
}

func NewApiGatewayRestApiEnumerator(repo repository.ApiGatewayRepository, factory resource.ResourceFactory) *ApiGatewayRestApiEnumerator {
	return &ApiGatewayRestApiEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayRestApiEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayRestApiResourceType
}

func (e *ApiGatewayRestApiEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllRestApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(apis))

	for _, api := range apis {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*api.Id,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

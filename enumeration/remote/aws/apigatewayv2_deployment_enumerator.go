package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type ApiGatewayV2DeploymentEnumerator struct {
	repository repository.ApiGatewayV2Repository
	factory    resource.ResourceFactory
}

func NewApiGatewayV2DeploymentEnumerator(repo repository.ApiGatewayV2Repository, factory resource.ResourceFactory) *ApiGatewayV2DeploymentEnumerator {
	return &ApiGatewayV2DeploymentEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *ApiGatewayV2DeploymentEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsApiGatewayV2DeploymentResourceType
}

func (e *ApiGatewayV2DeploymentEnumerator) Enumerate() ([]*resource.Resource, error) {
	apis, err := e.repository.ListAllApis()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsApiGatewayV2ApiResourceType)
	}

	var results []*resource.Resource
	for _, api := range apis {
		deployments, err := e.repository.ListAllApiDeployments(api.ApiId)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}
		for _, deployment := range deployments {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*deployment.DeploymentId,
					map[string]interface{}{},
				),
			)
		}
	}
	return results, err
}

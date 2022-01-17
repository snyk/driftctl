package aws

import "github.com/snyk/driftctl/pkg/resource"

const AwsApiGatewayV2MappingResourceType = "aws_apigatewayv2_api_mapping"

func initAwsApiGatewayV2MappingMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(
		AwsApiGatewayV2MappingResourceType,
		func(res *resource.Resource) map[string]string {
			return map[string]string{
				"Api":   *res.Attributes().GetString("api_id"),
				"Stage": *res.Attributes().GetString("stage"),
			}
		},
	)
}

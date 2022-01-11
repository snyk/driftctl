package aws

import "github.com/snyk/driftctl/pkg/resource"

const AwsApiGatewayV2ModelResourceType = "aws_apigatewayv2_model"

func initAwsApiGatewayV2ModelMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(
		AwsApiGatewayV2ModelResourceType,
		func(res *resource.Resource) map[string]string {
			return map[string]string{
				"name": *res.Attributes().GetString("name"),
			}
		},
	)
}

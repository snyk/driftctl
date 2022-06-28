package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsLambdaEventSourceMappingResourceType = "aws_lambda_event_source_mapping"

func initAwsLambdaEventSourceMappingMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsLambdaEventSourceMappingResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		source := val.GetString("event_source_arn")
		dest := val.GetString("function_name")
		if source != nil && *source != "" && dest != nil && *dest != "" {
			attrs["Source"] = *source
			attrs["Dest"] = *dest
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(AwsLambdaEventSourceMappingResourceType, resource.FlagDeepMode)
}

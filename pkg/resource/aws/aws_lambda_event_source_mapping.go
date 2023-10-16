package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsLambdaEventSourceMappingResourceType = "aws_lambda_event_source_mapping"

func initAwsLambdaEventSourceMappingMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsLambdaEventSourceMappingResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"state_transition_reason"})
		val.SafeDelete([]string{"state"})
		val.SafeDelete([]string{"starting_position_timestamp"})
		val.SafeDelete([]string{"starting_position"})
		val.SafeDelete([]string{"last_processing_result"})
		val.SafeDelete([]string{"last_modified"})
	})
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
}

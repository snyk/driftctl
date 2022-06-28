package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func initAwsLambdaEventSourceMappingMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(aws.AwsLambdaEventSourceMappingResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"state_transition_reason"})
		val.SafeDelete([]string{"state"})
		val.SafeDelete([]string{"starting_position_timestamp"})
		val.SafeDelete([]string{"starting_position"})
		val.SafeDelete([]string{"last_processing_result"})
		val.SafeDelete([]string{"last_modified"})
	})
}

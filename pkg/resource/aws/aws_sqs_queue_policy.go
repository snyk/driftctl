package aws

import (
	"github.com/snyk/driftctl/pkg/helpers"
	"github.com/snyk/driftctl/pkg/resource"
)

const AwsSqsQueuePolicyResourceType = "aws_sqs_queue_policy"

func initAwsSQSQueuePolicyMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.UpdateSchema(AwsSqsQueuePolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
	resourceSchemaRepository.SetNormalizeFunc(AwsSqsQueuePolicyResourceType, func(res *resource.Resource) {
		val := res.Attrs
		jsonString, err := helpers.NormalizeJsonString((*val)["policy"])
		if err != nil {
			return
		}
		_ = val.SafeSet([]string{"policy"}, jsonString)
	})
	resourceSchemaRepository.SetFlags(AwsSqsQueuePolicyResourceType, resource.FlagDeepMode)
}

package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/helpers"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsSqsQueuePolicyResourceType = "aws_sqs_queue_policy"

func initAwsSQSQueuePolicyMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsSqsQueuePolicyResourceType, func(res *resource.Resource) {
		val := res.Attrs
		jsonString, err := helpers.NormalizeJsonString((*val)["policy"])
		if err != nil {
			return
		}
		_ = val.SafeSet([]string{"policy"}, jsonString)
	})
	resourceSchemaRepository.UpdateSchema(AwsSqsQueuePolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
}

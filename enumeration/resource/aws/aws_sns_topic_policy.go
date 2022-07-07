package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsSnsTopicPolicyResourceType = "aws_sns_topic_policy"

func initSnsTopicPolicyMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsSnsTopicPolicyResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"topic_arn": res.ResourceId(),
		}
	})

	resourceSchemaRepository.UpdateSchema(AwsSnsTopicPolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
	resourceSchemaRepository.SetFlags(AwsSnsTopicPolicyResourceType, resource.FlagDeepMode)
}

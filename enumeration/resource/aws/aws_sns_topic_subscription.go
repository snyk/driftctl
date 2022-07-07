package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsSnsTopicSubscriptionResourceType = "aws_sns_topic_subscription"

func initSnsTopicSubscriptionMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsSnsTopicSubscriptionResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"SubscriptionId": res.ResourceId(),
		}
	})

	resourceSchemaRepository.UpdateSchema(AwsSnsTopicSubscriptionResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"delivery_policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
		"filter_policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})

	resourceSchemaRepository.SetFlags(AwsSnsTopicSubscriptionResourceType, resource.FlagDeepMode)
}

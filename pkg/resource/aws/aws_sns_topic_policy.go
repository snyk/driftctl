package aws

import (
	"github.com/cloudskiff/driftctl/pkg/helpers"
	"github.com/cloudskiff/driftctl/pkg/resource"
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

	resourceSchemaRepository.SetNormalizeFunc(AwsSnsTopicPolicyResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"owner"})
		jsonString, err := helpers.NormalizeJsonString((*val)["policy"])
		if err != nil {
			return
		}
		_ = val.SafeSet([]string{"policy"}, jsonString)
	})
}

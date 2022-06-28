package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsSnsTopicResourceType = "aws_sns_topic"

func initSnsTopicMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsSnsTopicResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"topic_arn": res.ResourceId(),
		}
	})
	resourceSchemaRepository.UpdateSchema(AwsSnsTopicResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"delivery_policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsSnsTopicResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if name := val.GetString("name"); name != nil && *name != "" {
			attrs["Name"] = *name
			if displayName := val.GetString("display_name"); displayName != nil && *displayName != "" {
				attrs["DisplayName"] = *displayName
			}
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(AwsSnsTopicResourceType, resource.FlagDeepMode)
}

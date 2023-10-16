package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsSnsTopicResourceType = "aws_sns_topic"

func initSnsTopicMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsSnsTopicResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.DeleteIfDefault("sqs_success_feedback_sample_rate")
		val.DeleteIfDefault("lambda_success_feedback_sample_rate")
		val.DeleteIfDefault("http_success_feedback_sample_rate")
		val.DeleteIfDefault("application_success_feedback_sample_rate")
		val.DeleteIfDefault("firehose_failure_feedback_role_arn")
		val.DeleteIfDefault("firehose_success_feedback_role_arn")
		val.SafeDelete([]string{"name_prefix"})
		val.SafeDelete([]string{"owner"})
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
}

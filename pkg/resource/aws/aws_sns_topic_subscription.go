package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/helpers"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsSnsTopicSubscriptionResourceType = "aws_sns_topic_subscription"

func initSnsTopicSubscriptionMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsSnsTopicSubscriptionResourceType, func(res *resource.Resource) {
		val := res.Attrs
		jsonString, err := helpers.NormalizeJsonString((*val)["delivery_policy"])
		if err == nil {
			_ = val.SafeSet([]string{"delivery_policy"}, jsonString)
		}

		jsonString, err = helpers.NormalizeJsonString((*val)["filter_policy"])
		if err == nil {
			_ = val.SafeSet([]string{"filter_policy"}, jsonString)
		}

		val.DeleteIfDefault("endpoint_auto_confirms")

		v, exists := val.Get("confirmation_timeout_in_minutes")
		if exists && v.(float64) == 1 {
			val.SafeDelete([]string{"confirmation_timeout_in_minutes"})
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
}

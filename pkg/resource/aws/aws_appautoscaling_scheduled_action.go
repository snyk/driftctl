package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsAppAutoscalingScheduledActionResourceType = "aws_appautoscaling_scheduled_action"

func initAwsAppAutoscalingScheduledActionMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsAppAutoscalingScheduledActionResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"name":               *res.Attributes().GetString("name"),
			"service_namespace":  *res.Attributes().GetString("service_namespace"),
			"scalable_dimension": *res.Attributes().GetString("scalable_dimension"),
			"resource_id":        *res.Attributes().GetString("resource_id"),
		}
	})
}

package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsAppAutoscalingPolicyResourceType = "aws_appautoscaling_policy"

func initAwsAppAutoscalingPolicyMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsAppAutoscalingPolicyResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"name":               *res.Attributes().GetString("name"),
			"resource_id":        *res.Attributes().GetString("resource_id"),
			"service_namespace":  *res.Attributes().GetString("service_namespace"),
			"scalable_dimension": *res.Attributes().GetString("scalable_dimension"),
		}
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsAppAutoscalingPolicyResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)
		if v := res.Attributes().GetString("scalable_dimension"); v != nil && *v != "" {
			attrs["Scalable dimension"] = *v
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(AwsAppAutoscalingPolicyResourceType, resource.FlagDeepMode)
}

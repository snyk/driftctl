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
}

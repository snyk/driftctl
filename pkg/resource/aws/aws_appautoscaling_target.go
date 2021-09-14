package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsAppAutoscalingTargetResourceType = "aws_appautoscaling_target"

func initAwsAppAutoscalingTargetMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsAppAutoscalingTargetResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"service_namespace":  *res.Attributes().GetString("service_namespace"),
			"scalable_dimension": *res.Attributes().GetString("scalable_dimension"),
		}
	})
}

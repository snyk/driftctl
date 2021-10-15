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
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsAppAutoscalingTargetResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)
		if v := res.Attributes().GetString("scalable_dimension"); v != nil && *v != "" {
			attrs["Scalable dimension"] = *v
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(AwsAppAutoscalingTargetResourceType, resource.FlagDeepMode)
	resourceSchemaRepository.SetDiscriminantFunc(AwsAppAutoscalingTargetResourceType, func(self, target *resource.Resource) bool {
		return *self.Attributes().GetString("scalable_dimension") == *target.Attributes().GetString("scalable_dimension")
	})
}

package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsAppAutoscalingTargetResourceType = "aws_appautoscaling_target"

func initAwsAppAutoscalingTargetMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsAppAutoscalingTargetResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)
		if v := res.Attributes().GetString("scalable_dimension"); v != nil && *v != "" {
			attrs["Scalable dimension"] = *v
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(AwsAppAutoscalingTargetResourceType, resource.FlagDeepMode)
}

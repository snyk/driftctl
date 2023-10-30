package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsAppAutoscalingTargetResourceType = "aws_appautoscaling_target"

func initAwsAppAutoscalingTargetMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsAppAutoscalingTargetResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)
		if v := res.Attributes().GetString("scalable_dimension"); v != nil && *v != "" {
			attrs["Scalable dimension"] = *v
		}
		return attrs
	})
	resourceSchemaRepository.SetDiscriminantFunc(AwsAppAutoscalingTargetResourceType, func(self, target *resource.Resource) bool {
		return *self.Attributes().GetString("scalable_dimension") == *target.Attributes().GetString("scalable_dimension")
	})
}

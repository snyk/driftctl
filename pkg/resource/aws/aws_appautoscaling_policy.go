package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsAppAutoscalingPolicyResourceType = "aws_appautoscaling_policy"

func initAwsAppAutoscalingPolicyMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsAppAutoscalingPolicyResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]string)
		if v := res.Attributes().GetString("scalable_dimension"); v != nil && *v != "" {
			attrs["Scalable dimension"] = *v
		}
		return attrs
	})
}

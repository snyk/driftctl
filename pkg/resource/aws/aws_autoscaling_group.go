package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsAutoScalingGroupResourceType = "aws_autoscaling_group"

func initAwsAutoScalingGroupMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsAutoScalingGroupResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"wait_for_capacity_timeout"})
		val.SafeDelete([]string{"timeouts"})
		val.SafeDelete([]string{"metrics_granularity"})
	})
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsAutoScalingGroupResourceType, func(res *resource.AbstractResource) map[string]string {
		return map[string]string{
			"id":   res.TerraformId(),
			"name": res.TerraformId(),
		}
	})
}

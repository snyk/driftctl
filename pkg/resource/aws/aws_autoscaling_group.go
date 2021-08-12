package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsAutoScalingGroupResourceType = "aws_autoscaling_group"

func initAwsAutoScalingGroupMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsEKSClusterResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"force_delete"})
		val.SafeDelete([]string{"force_delete_warm_pool"})
		val.SafeDelete([]string{"timeouts"})
		val.SafeDelete([]string{"wait_for_capacity_timeout"})
	})
}

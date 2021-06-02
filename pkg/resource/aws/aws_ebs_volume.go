package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsEbsVolumeResourceType = "aws_ebs_volume"

func initAwsEbsVolumeMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsEbsVolumeResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"arn"})
		val.SafeDelete([]string{"outpost_arn"})
		val.SafeDelete([]string{"snapshot_id"})
	})
}

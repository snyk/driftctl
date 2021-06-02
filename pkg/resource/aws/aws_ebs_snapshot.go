package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsEbsSnapshotResourceType = "aws_ebs_snapshot"

func initAwsEbsSnapshotMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsEbsSnapshotResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
}

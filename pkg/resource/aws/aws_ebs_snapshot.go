package aws

import (
	"github.com/snyk/driftctl/pkg/resource"
)

const AwsEbsSnapshotResourceType = "aws_ebs_snapshot"

func initAwsEbsSnapshotMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsEbsSnapshotResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetFlags(AwsEbsSnapshotResourceType, resource.FlagDeepMode)
}

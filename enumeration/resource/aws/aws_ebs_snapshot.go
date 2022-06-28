package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsEbsSnapshotResourceType = "aws_ebs_snapshot"

func initAwsEbsSnapshotMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsEbsSnapshotResourceType, resource.FlagDeepMode)
}

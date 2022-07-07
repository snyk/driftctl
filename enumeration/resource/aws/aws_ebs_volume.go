package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsEbsVolumeResourceType = "aws_ebs_volume"

func initAwsEbsVolumeMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsEbsVolumeResourceType, resource.FlagDeepMode)
}

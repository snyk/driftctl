package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func initAwsEbsVolumeMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(aws.AwsEbsVolumeResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"arn"})
		val.SafeDelete([]string{"outpost_arn"})
		val.SafeDelete([]string{"snapshot_id"})
		val.DeleteIfDefault("throughput")
	})
}

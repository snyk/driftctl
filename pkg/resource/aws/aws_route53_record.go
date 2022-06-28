package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func initAwsRoute53RecordMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(aws.AwsRoute53RecordResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.DeleteIfDefault("health_check_id")
		val.DeleteIfDefault("set_identifier")
		val.DeleteIfDefault("ttl")
		val.SafeDelete([]string{"name"})
		val.SafeDelete([]string{"allow_overwrite"})
	})
}

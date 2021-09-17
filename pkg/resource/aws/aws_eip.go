package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsEipResourceType = "aws_eip"

func initAwsEipMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsEipResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetFlags(AwsEipResourceType, resource.FlagDeepMode)
}

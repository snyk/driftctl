package aws

import (
	"github.com/snyk/driftctl/pkg/resource"
)

const AwsAmiResourceType = "aws_ami"

func initAwsAmiMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsAmiResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetFlags(AwsAmiResourceType, resource.FlagDeepMode)
}

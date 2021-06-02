package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsAmiResourceType = "aws_ami"

func initAwsAmiMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsAmiResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
}

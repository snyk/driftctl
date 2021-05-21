package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsKeyPairResourceType = "aws_key_pair"

func initAwsKeyPairMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsKeyPairResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"key_name_prefix"})
		val.SafeDelete([]string{"public_key"})
	})
}

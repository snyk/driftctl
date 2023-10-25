package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsAmiResourceType = "aws_ami"

func initAwsAmiMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsAmiResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
	})
}

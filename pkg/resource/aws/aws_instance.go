package aws

import (
	"github.com/hashicorp/go-version"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func initAwsInstanceMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(aws.AwsInstanceResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})

		if v, _ := version.NewVersion("3.38.0"); res.Schema().ProviderVersion.LessThan(v) {
			val.SafeDelete([]string{"instance_initiated_shutdown_behavior"})
		}
	})
}

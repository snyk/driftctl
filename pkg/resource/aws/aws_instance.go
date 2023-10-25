package aws

import (
	"github.com/hashicorp/go-version"
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsInstanceResourceType = "aws_instance"

func initAwsInstanceMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsInstanceResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})
		val.SafeDelete([]string{"arn"})

		if v, _ := version.NewVersion("3.38.0"); res.Schema().ProviderVersion.LessThan(v) {
			val.SafeDelete([]string{"instance_initiated_shutdown_behavior"})
		}
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsInstanceResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if tags := val.GetMap("tags"); tags != nil {
			if name, ok := tags["Name"]; ok {
				attrs["Name"] = name.(string)
			}
		}
		return attrs
	})
}

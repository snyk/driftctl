package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func initAwsCloudfrontDistributionMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(aws.AwsCloudfrontDistributionResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"etag"})
		val.SafeDelete([]string{"last_modified_time"})
		val.SafeDelete([]string{"retain_on_delete"})
		val.SafeDelete([]string{"status"})
		val.SafeDelete([]string{"wait_for_deployment"})
	})
}

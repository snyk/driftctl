package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsCloudfrontDistributionResourceType = "aws_cloudfront_distribution"

func initAwsCloudfrontDistributionMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsCloudfrontDistributionResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"etag"})
		val.SafeDelete([]string{"last_modified_time"})
		val.SafeDelete([]string{"retain_on_delete"})
		val.SafeDelete([]string{"status"})
		val.SafeDelete([]string{"wait_for_deployment"})
	})
}

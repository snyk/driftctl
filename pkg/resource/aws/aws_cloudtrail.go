package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsCloudtrailResourceType = "aws_cloudtrail"

func initAwsCloudtrailMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	/*resourceSchemaRepository.SetNormalizeFunc(AwsCloudfrontDistributionResourceType, func(res *resource.Resource) {
		val := res.Attrs
	})*/
	resourceSchemaRepository.SetFlags(AwsCloudfrontDistributionResourceType, resource.FlagDeepMode)

}

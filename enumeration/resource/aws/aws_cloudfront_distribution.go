package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsCloudfrontDistributionResourceType = "aws_cloudfront_distribution"

func initAwsCloudfrontDistributionMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsCloudfrontDistributionResourceType, resource.FlagDeepMode)
}

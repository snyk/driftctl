package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsS3BucketAnalyticsConfigurationResourceType = "aws_s3_bucket_analytics_configuration"

func initAwsS3BucketAnalyticsConfigurationMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsS3BucketAnalyticsConfigurationResourceType, resource.FlagDeepMode)
}

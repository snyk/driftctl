package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsS3BucketAnalyticsConfigurationResourceType = "aws_s3_bucket_analytics_configuration"

func initAwsS3BucketAnalyticsConfigurationMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsS3BucketAnalyticsConfigurationResourceType, resource.FlagDeepMode)
}

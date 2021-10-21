package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsS3BucketAnalyticsConfigurationResourceType = "aws_s3_bucket_analytics_configuration"

func initAwsS3BucketAnalyticsConfigurationMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsS3BucketAnalyticsConfigurationResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"alias": *res.Attributes().GetString("region"),
		}
	})
	resourceSchemaRepository.SetFlags(AwsS3BucketAnalyticsConfigurationResourceType, resource.FlagDeepMode)
}

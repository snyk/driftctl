package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsS3BucketMetricResourceType = "aws_s3_bucket_metric"

func initAwsS3BucketMetricMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsS3BucketMetricResourceType, resource.FlagDeepMode)
}

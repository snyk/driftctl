package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsS3BucketNotificationResourceType = "aws_s3_bucket_notification"

func initAwsS3BucketNotificationMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsS3BucketNotificationResourceType, resource.FlagDeepMode)
}

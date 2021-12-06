package aws

import "github.com/snyk/driftctl/pkg/resource"

const AwsS3BucketNotificationResourceType = "aws_s3_bucket_notification"

func initAwsS3BucketNotificationMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsS3BucketNotificationResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"alias": *res.Attributes().GetString("region"),
		}
	})
	resourceSchemaRepository.SetFlags(AwsS3BucketNotificationResourceType, resource.FlagDeepMode)
}

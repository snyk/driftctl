package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsS3BucketNotificationResourceType = "aws_s3_bucket_notification"

func initAwsS3BucketNotificationMetadata(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsS3BucketNotificationResourceType, resource.FlagDeepMode)
}

package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsS3BucketMetricResourceType = "aws_s3_bucket_metric"

func initAwsS3BucketMetricMetadata(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsS3BucketMetricResourceType, resource.FlagDeepMode)
}

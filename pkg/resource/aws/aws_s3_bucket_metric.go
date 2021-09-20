package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsS3BucketMetricResourceType = "aws_s3_bucket_metric"

func initAwsS3BucketMetricMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsS3BucketMetricResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"alias": *res.Attributes().GetString("region"),
		}
	})
	resourceSchemaRepository.SetFlags(AwsS3BucketMetricResourceType, resource.FlagDeepMode)
}

package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsS3BucketInventoryResourceType = "aws_s3_bucket_inventory"

func initAwsS3BucketInventoryMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsS3BucketInventoryResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"alias": *res.Attributes().GetString("region"),
		}
	})
	resourceSchemaRepository.SetFlags(AwsS3BucketInventoryResourceType, resource.FlagDeepMode)
}

package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsS3BucketInventoryResourceType = "aws_s3_bucket_inventory"

func initAwsS3BucketInventoryMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsS3BucketInventoryResourceType, resource.FlagDeepMode)
}

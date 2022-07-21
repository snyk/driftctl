package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsS3BucketInventoryResourceType = "aws_s3_bucket_inventory"

func initAwsS3BucketInventoryMetadata(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsS3BucketInventoryResourceType, resource.FlagDeepMode)
}

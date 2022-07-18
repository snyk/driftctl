package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsS3BucketPolicyResourceType = "aws_s3_bucket_policy"

func initAwsS3BucketPolicyMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.UpdateSchema(AwsS3BucketPolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
	resourceSchemaRepository.SetFlags(AwsS3BucketPolicyResourceType, resource.FlagDeepMode)
}

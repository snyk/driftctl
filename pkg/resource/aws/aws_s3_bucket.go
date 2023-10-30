package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsS3BucketResourceType = "aws_s3_bucket"

func initAwsS3BucketMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsS3BucketResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"force_destroy"})
		val.SafeDelete([]string{"bucket_prefix"})
	})
	resourceSchemaRepository.UpdateSchema(AwsS3BucketResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
}

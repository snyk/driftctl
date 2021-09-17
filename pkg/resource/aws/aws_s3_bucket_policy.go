package aws

import (
	"github.com/cloudskiff/driftctl/pkg/helpers"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsS3BucketPolicyResourceType = "aws_s3_bucket_policy"

func initAwsS3BucketPolicyMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsS3BucketPolicyResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"alias": *res.Attributes().GetString("region"),
		}
	})
	resourceSchemaRepository.UpdateSchema(AwsS3BucketPolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
	resourceSchemaRepository.SetNormalizeFunc(AwsS3BucketPolicyResourceType, func(res *resource.Resource) {
		val := res.Attrs
		jsonString, err := helpers.NormalizeJsonString((*val)["policy"])
		if err != nil {
			return
		}
		_ = val.SafeSet([]string{"policy"}, jsonString)
	})
	resourceSchemaRepository.SetFlags(AwsS3BucketPolicyResourceType, resource.FlagDeepMode)
}

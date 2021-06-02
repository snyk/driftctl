package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsIamUserPolicyResourceType = "aws_iam_user_policy"

func initAwsIAMUserPolicyMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.UpdateSchema(AwsIamUserPolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
}

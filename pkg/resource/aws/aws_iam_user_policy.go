package aws

import "github.com/snyk/driftctl/enumeration/resource"

const AwsIamUserPolicyResourceType = "aws_iam_user_policy"

func initAwsIAMUserPolicyMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.UpdateSchema(AwsIamUserPolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
	resourceSchemaRepository.SetFlags(AwsIamUserPolicyResourceType, resource.FlagDeepMode)
}

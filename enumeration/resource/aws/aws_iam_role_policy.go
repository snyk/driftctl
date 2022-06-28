package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsIamRolePolicyResourceType = "aws_iam_role_policy"

func initAwsIAMRolePolicyMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.UpdateSchema(AwsIamRolePolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
	resourceSchemaRepository.SetFlags(AwsIamRolePolicyResourceType, resource.FlagDeepMode)
}

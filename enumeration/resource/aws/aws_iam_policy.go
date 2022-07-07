package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsIamPolicyResourceType = "aws_iam_policy"

func initAwsIAMPolicyMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.UpdateSchema(AwsIamPolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
	resourceSchemaRepository.SetFlags(AwsIamPolicyResourceType, resource.FlagDeepMode)
}

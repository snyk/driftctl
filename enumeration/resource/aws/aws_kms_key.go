package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsKmsKeyResourceType = "aws_kms_key"

func initAwsKmsKeyMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.UpdateSchema(AwsKmsKeyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
	resourceSchemaRepository.SetFlags(AwsKmsKeyResourceType, resource.FlagDeepMode)
}

package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/helpers"
)

const AwsIamPolicyResourceType = "aws_iam_policy"

func initAwsIAMPolicyMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsIamPolicyResourceType, func(res *resource.Resource) {
		val := res.Attrs
		jsonString, err := helpers.NormalizeJsonString((*val)["policy"])
		if err == nil {
			_ = val.SafeSet([]string{"policy"}, jsonString)
		}

		val.SafeDelete([]string{"name_prefix"})
	})
	resourceSchemaRepository.UpdateSchema(AwsIamPolicyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
	resourceSchemaRepository.SetFlags(AwsIamPolicyResourceType, resource.FlagDeepMode)
}

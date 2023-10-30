package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/helpers"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsKmsKeyResourceType = "aws_kms_key"

func initAwsKmsKeyMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsKmsKeyResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"deletion_window_in_days"})
		jsonString, err := helpers.NormalizeJsonString((*val)["policy"])
		if err != nil {
			return
		}
		_ = val.SafeSet([]string{"policy"}, jsonString)
	})
	resourceSchemaRepository.UpdateSchema(AwsKmsKeyResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
}

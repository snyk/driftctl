package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsIamAccessKeyResourceType = "aws_iam_access_key"

func initAwsIAMAccessKeyMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsIamAccessKeyResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if user := val.GetString("user"); user != nil && *user != "" {
			attrs["User"] = *user
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(AwsIamAccessKeyResourceType, resource.FlagDeepMode)
}

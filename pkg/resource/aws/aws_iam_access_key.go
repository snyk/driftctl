package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsIamAccessKeyResourceType = "aws_iam_access_key"

func initAwsIAMAccessKeyMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsIamAccessKeyResourceType, func(res *resource.Resource) {
		val := res.Attrs
		// As we can't read secrets from aws API once access_key created we need to set
		// fields retrieved from state to nil to avoid drift
		// We can't detect drift if we cannot retrieve latest value from aws API for fields like secrets, passwords etc ...
		val.SafeDelete([]string{"secret"})
		val.SafeDelete([]string{"ses_smtp_password_v4"})
		val.SafeDelete([]string{"ses_smtp_password"})
		val.SafeDelete([]string{"encrypted_secret"})
		val.SafeDelete([]string{"key_fingerprint"})
		val.SafeDelete([]string{"pgp_key"})
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsIamAccessKeyResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if user := val.GetString("user"); user != nil && *user != "" {
			attrs["User"] = *user
		}
		return attrs
	})
}

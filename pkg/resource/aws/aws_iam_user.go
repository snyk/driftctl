package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsIamUserResourceType = "aws_iam_user"

func initAwsIAMUserMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsIamUserResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		permissionsBoundary, exist := val.Get("permissions_boundary")
		if exist && permissionsBoundary == "" {
			val.SafeDelete([]string{"permissions_boundary"})
		}
		val.SafeDelete([]string{"force_destroy"})
	})
}

package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func initAwsIAMUserMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(aws.AwsIamUserResourceType, func(res *resource.Resource) {
		val := res.Attrs
		permissionsBoundary, exist := val.Get("permissions_boundary")
		if exist && permissionsBoundary == "" {
			val.SafeDelete([]string{"permissions_boundary"})
		}
		val.SafeDelete([]string{"force_destroy"})
	})
}

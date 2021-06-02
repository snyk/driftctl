package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsDefaultSecurityGroupResourceType = "aws_default_security_group"

func initAwsDefaultSecurityGroupMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsDefaultSecurityGroupResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"revoke_rules_on_delete"})
		val.SafeDelete([]string{"ingress"})
		val.SafeDelete([]string{"egress"})
	})
}

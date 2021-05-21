package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsSecurityGroupResourceType = "aws_security_group"

func initAwsSecurityGroupMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsSecurityGroupResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"revoke_rules_on_delete"})
		val.SafeDelete([]string{"timeouts"})

		//TODO We need to find a way to warn users that some rules in their states could be unmanaged
		val.SafeDelete([]string{"ingress"})
		val.SafeDelete([]string{"egress"})
	})
}

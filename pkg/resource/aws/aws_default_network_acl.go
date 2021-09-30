package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsDefaultNetworkACLResourceType = "aws_default_network_acl"

func initAwsDefaultNetworkACLMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsDefaultNetworkACLResourceType, resource.FlagDeepMode)
	resourceSchemaRepository.SetNormalizeFunc(AwsDefaultNetworkACLResourceType, func(res *resource.Resource) {
		res.Attrs.SafeDelete([]string{"default_network_acl_id"})

		// https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/default_network_acl#managing-subnets-in-a-default-network-acl
		res.Attrs.SafeDelete([]string{"subnet_ids"})

	})
}

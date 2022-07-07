package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func initAwsDefaultNetworkACLMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(aws.AwsDefaultNetworkACLResourceType, func(res *resource.Resource) {
		res.Attrs.SafeDelete([]string{"default_network_acl_id"})

		// https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/default_network_acl#managing-subnets-in-a-default-network-acl
		res.Attrs.SafeDelete([]string{"subnet_ids"})

	})
}

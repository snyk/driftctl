package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsDefaultNetworkACLResourceType = "aws_default_network_acl"

func initAwsDefaultNetworkACLMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsDefaultNetworkACLResourceType, func(res *resource.Resource) {
		res.Attrs.SafeDelete([]string{"default_network_acl_id"})

		// https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/default_network_acl#managing-subnets-in-a-default-network-acl
		res.Attrs.SafeDelete([]string{"subnet_ids"})

	})

}

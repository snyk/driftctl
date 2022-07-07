package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func initAwsRouteMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(aws.AwsRouteResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"timeouts"})

		val.DeleteIfDefault("vpc_endpoint_id")
		val.DeleteIfDefault("local_gateway_id")
		val.DeleteIfDefault("destination_cidr_block")
		val.DeleteIfDefault("destination_ipv6_cidr_block")
		val.DeleteIfDefault("destination_prefix_list_id")
		val.DeleteIfDefault("egress_only_gateway_id")
		val.DeleteIfDefault("nat_gateway_id")
		val.DeleteIfDefault("instance_id")
		val.DeleteIfDefault("network_interface_id")
		val.DeleteIfDefault("transit_gateway_id")
		val.DeleteIfDefault("vpc_peering_connection_id")
		val.DeleteIfDefault("destination_prefix_list_id")
		val.DeleteIfDefault("instance_owner_id")
		val.DeleteIfDefault("carrier_gateway_id")
	})
}

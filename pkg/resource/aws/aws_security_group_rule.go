package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func initAwsSecurityGroupRuleMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(aws.AwsSecurityGroupRuleResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.DeleteIfDefault("security_group_id")
		val.DeleteIfDefault("source_security_group_id")

		// On first run, this field is set to null in state file and to "" after one refresh or apply
		// This ensure that if we find a nil value we dont drift
		val.DeleteIfDefault("description")

		// If protocol is all (e.g. -1), tcp, udp, icmp or icmpv6 then we leave the resource untouched
		// Else we delete the FromPort/ToPort and recreate the rule's id
		switch *val.GetString("protocol") {
		case "-1", "tcp", "udp", "icmp", "icmpv6":
			return
		}

		val.SafeDelete([]string{"from_port"})
		val.SafeDelete([]string{"to_port"})
		id := aws.CreateSecurityGroupRuleIdHash(val)
		_ = val.SafeSet([]string{"id"}, id)
		res.Id = id
	})
}

package aws

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsNetworkACLRuleResourceType = "aws_network_acl_rule"

func initAwsNetworkACLRuleMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsNetworkACLRuleResourceType, resource.FlagDeepMode)
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsNetworkACLRuleResourceType, func(res *resource.Resource) map[string]string {

		ruleNumber := strconv.FormatInt(int64(*res.Attrs.GetInt("rule_number")), 10)
		if ruleNumber == "32767" {
			ruleNumber = "*"
		}

		attrs := map[string]string{
			"Network":     *res.Attrs.GetString("network_acl_id"),
			"Egress":      strconv.FormatBool(*res.Attrs.GetBool("egress")),
			"Rule number": ruleNumber,
		}

		if proto := res.Attrs.GetString("protocol"); proto != nil {
			if *proto == "-1" {
				*proto = "All"
			}
			attrs["Protocol"] = *proto
		}

		if res.Attrs.GetFloat64("from_port") != nil && res.Attrs.GetFloat64("to_port") != nil {
			attrs["Port range"] = fmt.Sprintf("%d - %d",
				int64(*res.Attrs.GetFloat64("from_port")),
				int64(*res.Attrs.GetFloat64("to_port")),
			)
		}

		if cidr := res.Attrs.GetString("cidr_block"); cidr != nil && *cidr != "" {
			attrs["CIDR"] = *cidr
		}

		if cidr := res.Attrs.GetString("ipv6_cidr_block"); cidr != nil && *cidr != "" {
			attrs["CIDR"] = *cidr
		}

		return attrs
	})
}

func CreateNetworkACLRuleID(networkAclId string, ruleNumber int64, egress bool, protocol string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", networkAclId))
	buf.WriteString(fmt.Sprintf("%d-", ruleNumber))
	buf.WriteString(fmt.Sprintf("%t-", egress))
	buf.WriteString(fmt.Sprintf("%s-", protocol))
	return fmt.Sprintf("nacl-%d", hashcode.String(buf.String()))
}

package aws

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsSecurityGroupRuleResourceType = "aws_security_group_rule"

func CreateSecurityGroupRuleIdHash(attrs *resource.Attributes) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", *attrs.GetString("security_group_id")))
	if (*attrs)["from_port"] != nil && (*attrs)["from_port"].(int) > 0 {
		buf.WriteString(fmt.Sprintf("%d-", (*attrs)["from_port"].(int)))
	}
	if (*attrs)["to_port"] != nil && (*attrs)["to_port"].(int) > 0 {
		buf.WriteString(fmt.Sprintf("%d-", (*attrs)["to_port"].(int)))
	}
	buf.WriteString(fmt.Sprintf("%s-", *attrs.GetString("protocol")))
	buf.WriteString(fmt.Sprintf("%s-", *attrs.GetString("type")))

	if attrs.GetSlice("cidr_blocks") != nil {
		for _, v := range attrs.GetSlice("cidr_blocks") {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	if attrs.GetSlice("ipv6_cidr_blocks") != nil {
		for _, v := range attrs.GetSlice("ipv6_cidr_blocks") {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	if attrs.GetSlice("prefix_list_ids") != nil {
		for _, v := range attrs.GetSlice("prefix_list_ids") {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	if (attrs.GetBool("self") != nil && *attrs.GetBool("self")) ||
		(attrs.GetString("source_security_group_id") != nil && *attrs.GetString("source_security_group_id") != "") {
		if attrs.GetBool("self") != nil && *attrs.GetBool("self") {
			buf.WriteString(fmt.Sprintf("%s-", *attrs.GetString("security_group_id")))
		} else {
			buf.WriteString(fmt.Sprintf("%s-", *attrs.GetString("source_security_group_id")))
		}
		buf.WriteString("-")
	}

	return fmt.Sprintf("sgrule-%d", hashcode.String(buf.String()))
}

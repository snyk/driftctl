package aws

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/flatmap"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsSecurityGroupRuleResourceType = "aws_security_group_rule"

func CreateSecurityGroupRuleIdHash(attrs *resource.Attributes) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", *attrs.GetString("security_group_id")))
	if attrs.GetInt("from_port") != nil && *attrs.GetInt("from_port") > 0 {
		buf.WriteString(fmt.Sprintf("%d-", *attrs.GetInt("from_port")))
	}
	if attrs.GetInt("to_port") != nil && *attrs.GetInt("to_port") > 0 {
		buf.WriteString(fmt.Sprintf("%d-", *attrs.GetInt("to_port")))
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

func initAwsSecurityGroupRuleMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsSecurityGroupRuleResourceType, func(res *resource.Resource) map[string]string {
		attrs := make(map[string]interface{})
		if v, ok := res.Attributes().Get("type"); ok {
			attrs["type"] = v
		}
		if v, ok := res.Attributes().Get("protocol"); ok {
			attrs["protocol"] = v
		}
		if v := res.Attributes().GetInt("from_port"); v != nil {
			attrs["from_port"] = *v
		}
		if v := res.Attributes().GetInt("to_port"); v != nil {
			attrs["to_port"] = *v
		}
		if v, ok := res.Attributes().Get("security_group_id"); ok {
			attrs["security_group_id"] = v
		}
		if v, ok := res.Attributes().Get("self"); ok {
			attrs["self"] = v
		}
		if v, ok := res.Attributes().Get("cidr_blocks"); ok {
			attrs["cidr_blocks"] = v
		}
		if v, ok := res.Attributes().Get("ipv6_cidr_blocks"); ok {
			attrs["ipv6_cidr_blocks"] = v
		}
		if v, ok := res.Attributes().Get("prefix_list_ids"); ok {
			attrs["prefix_list_ids"] = v
		}
		if v, ok := res.Attributes().Get("source_security_group_id"); ok {
			attrs["source_security_group_id"] = v
		}
		return flatmap.Flatten(attrs)
	})
	resourceSchemaRepository.SetHumanReadableAttributesFunc(AwsSecurityGroupRuleResourceType, func(res *resource.Resource) map[string]string {
		val := res.Attrs
		attrs := make(map[string]string)
		if sgID := val.GetString("security_group_id"); sgID != nil && *sgID != "" {
			attrs["SecurityGroup"] = *sgID
		}
		if protocol := val.GetString("protocol"); protocol != nil && *protocol != "" {
			if *protocol == "-1" {
				*protocol = "All"
			}
			attrs["Protocol"] = *protocol
		}
		fromPort := val.GetInt("from_port")
		toPort := val.GetInt("to_port")
		if fromPort != nil && toPort != nil {
			portRange := "All"
			if *fromPort != 0 && *fromPort == *toPort {
				portRange = fmt.Sprintf("%d", *fromPort)
			}
			if *fromPort != 0 && *toPort != 0 && *fromPort != *toPort {
				portRange = fmt.Sprintf("%d-%d", *fromPort, *toPort)
			}
			attrs["Ports"] = portRange
		}
		ty := val.GetString("type")
		if ty != nil && *ty != "" {
			attrs["Type"] = *ty
			var sourceOrDestination string
			switch *ty {
			case "egress":
				sourceOrDestination = "Destination"
			case "ingress":
				sourceOrDestination = "Source"
			}
			if ipv4 := val.GetSlice("cidr_blocks"); len(ipv4) > 0 {
				attrs[sourceOrDestination] = join(ipv4, ", ")
			}
			if ipv6 := val.GetSlice("ipv6_cidr_blocks"); len(ipv6) > 0 {
				attrs[sourceOrDestination] = join(ipv6, ", ")
			}
			if prefixList := val.GetSlice("prefix_list_ids"); len(prefixList) > 0 {
				attrs[sourceOrDestination] = join(prefixList, ", ")
			}
			if sourceSgID := val.GetString("source_security_group_id"); sourceSgID != nil && *sourceSgID != "" {
				attrs[sourceOrDestination] = *sourceSgID
			}
		}
		return attrs
	})
	resourceSchemaRepository.SetFlags(AwsSecurityGroupRuleResourceType, resource.FlagDeepMode)
}

func join(elems []interface{}, sep string) string {
	firstElemt, ok := elems[0].(string)
	if !ok {
		panic("cannot join a slice that contains something else than strings")
	}
	switch len(elems) {
	case 0:
		return ""
	case 1:

		return firstElemt
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i].(string))
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(firstElemt)
	for _, s := range elems[1:] {
		b.WriteString(sep)
		elem, ok := s.(string)
		if !ok {
			panic("cannot join a slice that contains something else than strings")
		}
		b.WriteString(elem)
	}
	return b.String()
}

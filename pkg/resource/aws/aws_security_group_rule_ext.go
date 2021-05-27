package aws

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/hashicorp/terraform/helper/hashcode"
)

func (r *AwsSecurityGroupRule) NormalizeForState() (resource.Resource, error) {
	// On first run, this field is set to null in state file and to "" after one refresh or apply
	// This ensure that if we find a nil value we dont drift
	if r.Description == nil {
		r.Description = aws.String("")
	}

	// If protocol is all (e.g. -1), tcp, udp, icmp or icmpv6 then we return the resource
	// Else we change the FromPort/ToPort to 0 and recreate the rule's id
	switch *r.Protocol {
	case "-1", "tcp", "udp", "icmp", "icmpv6":
		return r, nil
	}

	r.FromPort = aws.Int(0)
	r.ToPort = aws.Int(0)
	r.Id = r.CreateIdHash()

	return r, nil
}

func (r *AwsSecurityGroupRule) NormalizeForProvider() (resource.Resource, error) {
	// When crafting the rule, by default terraform set it to null but not us
	// This ensure that if we find a nil value we dont drift
	if r.SourceSecurityGroupId != nil && *r.SourceSecurityGroupId == "" {
		r.SourceSecurityGroupId = nil
	}

	return r, nil
}

func (r *AwsSecurityGroupRule) CreateIdHash() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", *r.SecurityGroupId))
	if r.FromPort != nil && *r.FromPort > 0 {
		buf.WriteString(fmt.Sprintf("%d-", *r.FromPort))
	}
	if r.ToPort != nil && *r.ToPort > 0 {
		buf.WriteString(fmt.Sprintf("%d-", *r.ToPort))
	}
	buf.WriteString(fmt.Sprintf("%s-", *r.Protocol))
	buf.WriteString(fmt.Sprintf("%s-", *r.Type))

	if r.CidrBlocks != nil {
		for _, v := range *r.CidrBlocks {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	if r.Ipv6CidrBlocks != nil {
		for _, v := range *r.Ipv6CidrBlocks {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	if r.PrefixListIds != nil {
		for _, v := range *r.PrefixListIds {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	if (r.Self != nil && *r.Self) ||
		(r.SourceSecurityGroupId != nil && *r.SourceSecurityGroupId != "") {
		if r.Self != nil && *r.Self {
			buf.WriteString(fmt.Sprintf("%s-", *r.SecurityGroupId))
		} else {
			buf.WriteString(fmt.Sprintf("%s-", *r.SourceSecurityGroupId))
		}
		buf.WriteString("-")
	}

	return fmt.Sprintf("sgrule-%d", hashcode.String(buf.String()))
}

func (r *AwsSecurityGroupRule) Attributes() map[string]string {
	attrs := make(map[string]string)
	if r.Type != nil && *r.Type != "" {
		attrs["Type"] = *r.Type
	}
	if r.SecurityGroupId != nil && *r.SecurityGroupId != "" {
		attrs["SecurityGroup"] = *r.SecurityGroupId
	}
	if r.Protocol != nil && *r.Protocol != "" {
		proto := *r.Protocol
		if proto == "-1" {
			proto = "All"
		}
		attrs["Protocol"] = proto
	}
	if r.FromPort != nil && r.ToPort != nil {
		portRange := "All"
		if *r.FromPort != 0 && *r.FromPort == *r.ToPort {
			portRange = fmt.Sprintf("%d", *r.FromPort)
		}
		if *r.FromPort != 0 && *r.ToPort != 0 && *r.FromPort != *r.ToPort {
			portRange = fmt.Sprintf("%d-%d", *r.FromPort, *r.ToPort)
		}
		attrs["Ports"] = portRange
	}
	sourceOrDestination := "Source"
	if r.Type != nil && *r.Type == "egress" {
		sourceOrDestination = "Destination"
	}
	if r.CidrBlocks != nil && len(*r.CidrBlocks) > 0 {
		attrs[sourceOrDestination] = strings.Join(*r.CidrBlocks, ", ")
	}
	if r.Ipv6CidrBlocks != nil && len(*r.Ipv6CidrBlocks) > 0 {
		attrs[sourceOrDestination] = strings.Join(*r.Ipv6CidrBlocks, ", ")
	}
	if r.SourceSecurityGroupId != nil && *r.SourceSecurityGroupId != "" {
		attrs[sourceOrDestination] = *r.SecurityGroupId
	}
	if r.PrefixListIds != nil && len(*r.PrefixListIds) > 0 {
		attrs[sourceOrDestination] = strings.Join(*r.PrefixListIds, ", ")
	}
	return attrs
}

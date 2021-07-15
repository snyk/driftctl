package aws

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type VPCSecurityGroupRuleDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewVPCSecurityGroupRuleDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *VPCSecurityGroupRuleDetailsFetcher {
	return &VPCSecurityGroupRuleDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *VPCSecurityGroupRuleDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	attrs := make(map[string]string)

	if v, ok := res.Attributes().Get("type"); ok {
		attrs["type"] = fmt.Sprintf("%s", v)
	}
	if v, ok := res.Attributes().Get("protocol"); ok {
		attrs["protocol"] = fmt.Sprintf("%s", v)
	}
	if v, ok := res.Attributes().Get("from_port"); ok {
		attrs["from_port"] = fmt.Sprintf("%d", int(v.(float64)))
	}
	if v, ok := res.Attributes().Get("to_port"); ok {
		attrs["to_port"] = fmt.Sprintf("%d", int(v.(float64)))
	}
	if v, ok := res.Attributes().Get("security_group_id"); ok {
		attrs["security_group_id"] = fmt.Sprintf("%s", v)
	}
	if v, ok := res.Attributes().Get("self"); ok {
		attrs["self"] = fmt.Sprintf("%t", v)
	}
	if v, ok := res.Attributes().Get("cidr_blocks"); ok {
		attrs["cidr_blocks"] = fmt.Sprintf("%s", v)
	}
	if v, ok := res.Attributes().Get("ipv6_cidr_blocks"); ok {
		attrs["ipv6_cidr_blocks"] = fmt.Sprintf("%s", v)
	}
	if v, ok := res.Attributes().Get("prefix_list_ids"); ok {
		attrs["prefix_list_ids"] = fmt.Sprintf("%s", v)
	}
	if v, ok := res.Attributes().Get("source_security_group_id"); ok {
		attrs["source_security_group_id"] = fmt.Sprintf("%s", v)
	}

	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty:         aws.AwsSecurityGroupRuleResourceType,
		ID:         res.TerraformId(),
		Attributes: attrs,
	})
	if err != nil {
		return nil, err
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsSecurityGroupRuleResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}

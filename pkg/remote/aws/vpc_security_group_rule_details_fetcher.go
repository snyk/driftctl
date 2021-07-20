package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/hashicorp/terraform/flatmap"
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

	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty:         aws.AwsSecurityGroupRuleResourceType,
		ID:         res.TerraformId(),
		Attributes: flatmap.Flatten(attrs),
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

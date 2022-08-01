package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type EC2NetworkACLRuleEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2NetworkACLRuleEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2NetworkACLRuleEnumerator {
	return &EC2NetworkACLRuleEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2NetworkACLRuleEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsNetworkACLRuleResourceType
}

func (e *EC2NetworkACLRuleEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.ListAllNetworkACLs()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsNetworkACLResourceType)
	}

	results := make([]*resource.Resource, 0, len(resources))

	for _, res := range resources {
		for _, entry := range res.Entries {

			attrs := map[string]interface{}{
				"egress":         *entry.Egress,
				"network_acl_id": *res.NetworkAclId,
				"rule_action":    *entry.RuleAction, // Used in default middleware
				"rule_number":    *entry.RuleNumber, // Used in default middleware
				"protocol":       *entry.Protocol,   // Used in default middleware
			}

			if entry.CidrBlock != nil {
				attrs["cidr_block"] = *entry.CidrBlock
			}

			if entry.Ipv6CidrBlock != nil {
				attrs["ipv6_cidr_block"] = *entry.Ipv6CidrBlock
			}

			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					aws.CreateNetworkACLRuleID(
						*res.NetworkAclId,
						*entry.RuleNumber,
						*entry.Egress,
						*entry.Protocol,
					),
					attrs,
				),
			)
		}
	}

	return results, err
}

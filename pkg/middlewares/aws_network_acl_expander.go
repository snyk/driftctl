package middlewares

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// This middelware goal is to explode aws_network_acl ingress and egress block into a set of aws_network_acl_rule
type AwsNetworkACLExpander struct {
	resourceFactory resource.ResourceFactory
}

func NewAwsNetworkACLExpander(resourceFactory resource.ResourceFactory) AwsNetworkACLExpander {
	return AwsNetworkACLExpander{resourceFactory}
}

func (m AwsNetworkACLExpander) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newResourcesFromState := make([]*resource.Resource, 0, len(*resourcesFromState))

	for _, stateResource := range *resourcesFromState {
		// Ignore all resources other than network acl
		if stateResource.ResourceType() != aws.AwsNetworkACLResourceType &&
			stateResource.ResourceType() != aws.AwsDefaultNetworkACLResourceType {
			newResourcesFromState = append(newResourcesFromState, stateResource)
			continue
		}

		newResourcesFromState = append(newResourcesFromState, m.expandBlock(
			resourcesFromState,
			stateResource.ResourceId(),
			false,
			stateResource.Attrs.GetSlice("ingress"),
		)...)
		stateResource.Attrs.SafeDelete([]string{"ingress"})

		newResourcesFromState = append(newResourcesFromState, m.expandBlock(
			resourcesFromState,
			stateResource.ResourceId(),
			true,
			stateResource.Attrs.GetSlice("egress"),
		)...)
		stateResource.Attrs.SafeDelete([]string{"egress"})

		newResourcesFromState = append(newResourcesFromState, stateResource)
	}

	// Then we need to remove ingress and egress block from remote resource too
	newRemoteResources := make([]*resource.Resource, 0, len(*remoteResources))
	for _, remoteResource := range *remoteResources {
		if remoteResource.ResourceType() != aws.AwsNetworkACLResourceType &&
			remoteResource.ResourceType() != aws.AwsDefaultNetworkACLResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		remoteResource.Attrs.SafeDelete([]string{"ingress"})
		remoteResource.Attrs.SafeDelete([]string{"egress"})

		newRemoteResources = append(newRemoteResources, remoteResource)
	}

	*resourcesFromState = newResourcesFromState
	*remoteResources = newRemoteResources

	return nil
}

func (e *AwsNetworkACLExpander) expandBlock(resourcesFromState *[]*resource.Resource, networkAclId string, egress bool, ruleBlock []interface{}) []*resource.Resource {
	results := make([]*resource.Resource, 0, len(ruleBlock))

	for _, rule := range ruleBlock {
		attrs := rule.(map[string]interface{})

		attrs["rule_number"] = attrs["rule_no"]
		delete(attrs, "rule_no")

		attrs["egress"] = egress

		attrs["network_acl_id"] = networkAclId

		attrs["rule_action"] = attrs["action"]
		delete(attrs, "action")

		res := e.resourceFactory.CreateAbstractResource(
			aws.AwsNetworkACLRuleResourceType,
			aws.CreateNetworkACLRuleID(
				networkAclId,
				int64(attrs["rule_number"].(int)),
				egress,
				attrs["protocol"].(string),
			),
			attrs,
		)

		existInState := false
		for _, stateRes := range *resourcesFromState {
			if stateRes.Equal(res) {
				existInState = true
				break
			}
		}

		if !existInState {
			results = append(results, res)
		}
	}

	return results
}

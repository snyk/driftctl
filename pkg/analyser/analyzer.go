package analyser

import (
	"github.com/snyk/driftctl/enumeration/alerter"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/pkg/filter"

	"github.com/snyk/driftctl/enumeration/resource"
)

type UnmanagedSecurityGroupRulesAlert struct{}

func newUnmanagedSecurityGroupRulesAlert() *UnmanagedSecurityGroupRulesAlert {
	return &UnmanagedSecurityGroupRulesAlert{}
}

func (u *UnmanagedSecurityGroupRulesAlert) Message() string {
	return "You have unmanaged security group rules that could be false positives, find out more at https://docs.driftctl.com/limitations"
}

func (u *UnmanagedSecurityGroupRulesAlert) ShouldIgnoreResource() bool {
	return false
}

func (u *UnmanagedSecurityGroupRulesAlert) Resource() *resource.Resource {
	return nil
}

type ComputedDiffAlert struct{}

func NewComputedDiffAlert() *ComputedDiffAlert {
	return &ComputedDiffAlert{}
}

func (c *ComputedDiffAlert) Message() string {
	return "You have diffs on computed fields, check the documentation for potential false positive drifts: https://docs.driftctl.com/limitations"
}

func (c *ComputedDiffAlert) ShouldIgnoreResource() bool {
	return false
}

func (c *ComputedDiffAlert) Resource() *resource.Resource {
	return nil
}

type Analyzer struct {
	alerter *alerter.Alerter
	filter  filter.Filter
}

func NewAnalyzer(alerter *alerter.Alerter, filter filter.Filter) *Analyzer {
	return &Analyzer{alerter, filter}
}

func (a Analyzer) Analyze(remoteResources, resourcesFromState []*resource.Resource) (Analysis, error) {
	analysis := Analysis{}

	// Iterate on remote resources and filter ignored resources
	filteredRemoteResource := make([]*resource.Resource, 0, len(remoteResources))
	for _, remoteRes := range remoteResources {
		if a.filter.IsResourceIgnored(remoteRes) || a.alerter.IsResourceIgnored(remoteRes) {
			continue
		}
		filteredRemoteResource = append(filteredRemoteResource, remoteRes)
	}

	haveComputedDiff := false
	for _, stateRes := range resourcesFromState {
		i, _, found := findCorrespondingRes(filteredRemoteResource, stateRes)

		if a.filter.IsResourceIgnored(stateRes) || a.alerter.IsResourceIgnored(stateRes) {
			continue
		}

		if !found {
			analysis.AddDeleted(stateRes)
			continue
		}

		// Remove managed resources, so it will remain only unmanaged ones
		filteredRemoteResource = removeResourceByIndex(i, filteredRemoteResource)
		analysis.AddManaged(stateRes)
	}

	if a.hasUnmanagedSecurityGroupRules(filteredRemoteResource) {
		a.alerter.SendAlert("", newUnmanagedSecurityGroupRulesAlert())
	}

	if haveComputedDiff {
		a.alerter.SendAlert("", NewComputedDiffAlert())
	}

	// Add remaining unmanaged resources
	analysis.AddUnmanaged(filteredRemoteResource...)

	// Sort resources by Terraform Id
	// The purpose is to have a predictable output
	analysis.SortResources()

	analysis.SetAlerts(a.alerter.Retrieve())

	return analysis, nil
}

func findCorrespondingRes(resources []*resource.Resource, res *resource.Resource) (int, *resource.Resource, bool) {
	for i, r := range resources {
		if res.Equal(r) {
			return i, r, true
		}
	}
	return -1, nil, false
}

func removeResourceByIndex(i int, resources []*resource.Resource) []*resource.Resource {
	if i == len(resources)-1 {
		return resources[:len(resources)-1]
	}
	return append(resources[:i], resources[i+1:]...)
}

// hasUnmanagedSecurityGroupRules returns true if we find at least one unmanaged
// security group rule
func (a Analyzer) hasUnmanagedSecurityGroupRules(unmanagedResources []*resource.Resource) bool {
	for _, res := range unmanagedResources {
		if res.ResourceType() == resourceaws.AwsSecurityGroupRuleResourceType {
			return true
		}
	}
	return false
}

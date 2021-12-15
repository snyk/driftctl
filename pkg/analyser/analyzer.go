package analyser

import (
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/pkg/filter"
	resourceaws "github.com/snyk/driftctl/pkg/resource/aws"

	"github.com/snyk/driftctl/pkg/alerter"
	"github.com/snyk/driftctl/pkg/resource"
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

type AnalyzerOptions struct {
	Deep bool
}

type Analyzer struct {
	alerter *alerter.Alerter
	options AnalyzerOptions
	filter  filter.Filter
}

func NewAnalyzer(alerter *alerter.Alerter, options AnalyzerOptions, filter filter.Filter) *Analyzer {
	return &Analyzer{alerter, options, filter}
}

func (a Analyzer) CompareEnumeration(analysis *Analysis, remoteResources, resourcesFromState []*resource.Resource) *Analysis {
	// Iterate on remote resources and filter ignored resources
	filteredRemoteResources := make([]*resource.Resource, 0, len(remoteResources))
	for _, remoteRes := range remoteResources {
		if a.filter.IsResourceIgnored(remoteRes) || a.alerter.IsResourceIgnored(remoteRes) {
			continue
		}
		filteredRemoteResources = append(filteredRemoteResources, remoteRes)
	}

	for _, stateRes := range resourcesFromState {
		if a.filter.IsResourceIgnored(stateRes) || a.alerter.IsResourceIgnored(stateRes) {
			continue
		}

		i, _, found := resource.FindCorrespondingRes(filteredRemoteResources, stateRes)
		if !found {
			analysis.AddDeleted(stateRes)
			continue
		}

		// Remove managed resources, so it will remain only unmanaged ones
		filteredRemoteResources = removeResourceByIndex(i, filteredRemoteResources)
		analysis.AddManaged(stateRes)
	}

	if a.hasUnmanagedSecurityGroupRules(filteredRemoteResources) {
		a.alerter.SendAlert("", newUnmanagedSecurityGroupRulesAlert())
	}

	// Add remaining unmanaged resources
	analysis.AddUnmanaged(filteredRemoteResources...)

	return analysis
}

func (a Analyzer) CompleteAnalysis(analysis *Analysis, managedResources, resourcesFromState []*resource.Resource) *Analysis {
	// Stop there if we are not in deep mode, we do not want to compute diffs
	if !a.options.Deep {
		a.setAlerts(analysis)
		return analysis
	}

	haveComputedDiff := false
	for _, remoteRes := range managedResources {
		if a.filter.IsResourceIgnored(remoteRes) || a.alerter.IsResourceIgnored(remoteRes) {
			continue
		}

		_, stateRes, found := resource.FindCorrespondingRes(resourcesFromState, remoteRes)
		if !found {
			continue
		}

		// Stop if the resource is not compatible with deep mode
		if stateRes.Schema() != nil && !stateRes.Schema().Flags.HasFlag(resource.FlagDeepMode) {
			continue
		}

		var delta diff.Changelog
		delta, _ = diff.Diff(stateRes.Attributes(), remoteRes.Attributes())

		if len(delta) == 0 {
			continue
		}

		changelog := make([]Change, 0, len(delta))
		for _, change := range delta {
			if a.filter.IsFieldIgnored(stateRes, change.Path) {
				continue
			}
			c := Change{Change: change}
			resSchema := stateRes.Schema()
			if resSchema != nil {
				c.Computed = resSchema.IsComputedField(c.Path)
				c.JsonString = resSchema.IsJsonStringField(c.Path)
			}
			if c.Computed {
				haveComputedDiff = true
			}
			changelog = append(changelog, c)
		}
		if len(changelog) > 0 {
			analysis.AddDifference(Difference{
				Res:       stateRes,
				Changelog: changelog,
			})
		}
	}

	if haveComputedDiff {
		a.alerter.SendAlert("", NewComputedDiffAlert())
	}

	a.setAlerts(analysis)

	return analysis
}

func (a Analyzer) setAlerts(analysis *Analysis) {
	analysis.SetAlerts(a.alerter.Retrieve())
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

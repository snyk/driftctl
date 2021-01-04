package analyser

import (
	"sort"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/r3labs/diff/v2"
)

type Analyzer struct{}

type Filter interface {
	IsResourceIgnored(res resource.Resource) bool
	IsFieldIgnored(res resource.Resource, path []string) bool
}

func NewAnalyzer() Analyzer {
	return Analyzer{}
}

func (a Analyzer) Analyze(remoteResources []resource.Resource, resourcesFromState []resource.Resource, filter Filter) (Analysis, error) {
	analysis := Analysis{}

	// Iterate on remote resources and filter ignored resources
	filteredRemoteResource := make([]resource.Resource, 0, len(remoteResources))
	for _, remoteRes := range remoteResources {
		if filter.IsResourceIgnored(remoteRes) {
			continue
		}
		filteredRemoteResource = append(filteredRemoteResource, remoteRes)
	}

	for _, stateRes := range resourcesFromState {
		i, remoteRes, found := findCorrespondingRes(filteredRemoteResource, stateRes)

		if filter.IsResourceIgnored(stateRes) {
			continue
		}

		if !found {
			analysis.AddDeleted(stateRes)
			continue
		}

		// Remove managed resources, so it will remain only unmanaged ones
		filteredRemoteResource = removeResourceByIndex(i, filteredRemoteResource)
		analysis.AddManaged(stateRes)

		delta, _ := diff.Diff(stateRes, remoteRes)
		if len(delta) > 0 {
			sort.Slice(delta, func(i, j int) bool {
				return delta[i].Type < delta[j].Type
			})
			changelog := make([]diff.Change, 0, len(delta))
			for _, change := range delta {
				if filter.IsFieldIgnored(stateRes, change.Path) {
					continue
				}
				changelog = append(changelog, change)
			}
			if len(changelog) > 0 {
				analysis.AddDifference(Difference{
					Res:       stateRes,
					Changelog: changelog,
				})
			}
		}
	}

	// Add remaining unmanaged resources
	analysis.AddUnmanaged(filteredRemoteResource...)

	return analysis, nil
}

func findCorrespondingRes(resources []resource.Resource, res resource.Resource) (int, resource.Resource, bool) {
	for i, r := range resources {
		if resource.IsSameResource(res, r) {
			return i, r, true
		}
	}
	return -1, nil, false
}

func removeResourceByIndex(i int, resources []resource.Resource) []resource.Resource {
	if i == len(resources)-1 {
		return resources[:len(resources)-1]
	}
	return append(resources[:i], resources[i+1:]...)
}

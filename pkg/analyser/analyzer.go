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

	for _, stateRes := range resourcesFromState {
		i, remoteRes, found := findCorrespondingRes(remoteResources, stateRes)

		if filter.IsResourceIgnored(stateRes) {
			if found {
				remoteResources = append(remoteResources[:i], remoteResources[i+1:]...)
			}
			continue
		}

		if !found {
			analysis.AddDeleted(stateRes)
			continue
		}
		remoteResources = append(remoteResources[:i], remoteResources[i+1:]...)
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
	analysis.AddUnmanaged(remoteResources...)
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

package analyser

import (
	"reflect"
	"sort"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/r3labs/diff/v2"
)

type Analyzer struct {
	alerter *alerter.Alerter
}

type Filter interface {
	IsResourceIgnored(res resource.Resource) bool
	IsFieldIgnored(res resource.Resource, path []string) bool
}

func NewAnalyzer(alerter *alerter.Alerter) Analyzer {
	return Analyzer{alerter}
}

func (a Analyzer) Analyze(remoteResources, resourcesFromState []resource.Resource, filter Filter) (Analysis, error) {
	analysis := Analysis{}

	// Iterate on remote resources and filter ignored resources
	filteredRemoteResource := make([]resource.Resource, 0, len(remoteResources))
	for _, remoteRes := range remoteResources {
		if filter.IsResourceIgnored(remoteRes) || a.alerter.IsResourceIgnored(remoteRes) {
			continue
		}
		filteredRemoteResource = append(filteredRemoteResource, remoteRes)
	}

	haveComputedDiff := false
	for _, stateRes := range resourcesFromState {
		i, remoteRes, found := findCorrespondingRes(filteredRemoteResource, stateRes)

		if filter.IsResourceIgnored(stateRes) || a.alerter.IsResourceIgnored(stateRes) {
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
			changelog := make([]Change, 0, len(delta))
			for _, change := range delta {
				if filter.IsFieldIgnored(stateRes, change.Path) {
					continue
				}
				c := Change{Change: change}
				c.Computed = a.isComputedField(stateRes, c)
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
	}

	if haveComputedDiff {
		a.alerter.SendAlert("",
			alerter.Alert{
				Message: "You have diffs on computed fields, check the documentation for potential false positive drifts",
			})
	}

	// Add remaining unmanaged resources
	analysis.AddUnmanaged(filteredRemoteResource...)

	analysis.SetAlerts(a.alerter.Retrieve())

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

// isComputedField returns true if the field that generated the diff of a resource
// has a computed tag
func (a Analyzer) isComputedField(stateRes resource.Resource, change Change) bool {
	if field, ok := a.getField(reflect.TypeOf(stateRes), change.Path); ok {
		return field.Tag.Get("computed") == "true"
	}
	return false
}

// getField recursively finds the deepest field inside a resource depending on
// its path and its type
func (a Analyzer) getField(t reflect.Type, path []string) (reflect.StructField, bool) {
	switch t.Kind() {
	case reflect.Ptr:
		return a.getField(t.Elem(), path)
	case reflect.Slice:
		return a.getField(t.Elem(), path[1:])
	default:
		{
			if field, ok := t.FieldByName(path[0]); ok && a.hasNestedFields(field.Type) && len(path) > 1 {
				return a.getField(field.Type, path[1:])
			} else {
				return field, ok
			}
		}
	}
}

// hasNestedFields will return true if the current field is either a struct
// or a slice of struct
func (a Analyzer) hasNestedFields(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr:
		return a.hasNestedFields(t.Elem())
	case reflect.Slice:
		return t.Elem().Kind() == reflect.Struct
	default:
		return t.Kind() == reflect.Struct
	}
}

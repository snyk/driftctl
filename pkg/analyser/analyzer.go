package analyser

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

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
				if a.isComputedField(stateRes, c) {
					c.Computed = true
				}
				changelog = append(changelog, c)
			}
			if len(changelog) > 0 {
				analysis.AddDifference(Difference{
					Res:       stateRes,
					Changelog: changelog,
				})
				a.sendAlertOnComputedField(stateRes, changelog)
			}
		}
	}

	// Add remaining unmanaged resources
	analysis.AddUnmanaged(filteredRemoteResource...)

	analysis.AddAlerts(a.alerter.GetAlerts())

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

// sendAlertOnComputedField will send an alert to a channel for diffs on computed field
func (a Analyzer) sendAlertOnComputedField(stateRes resource.Resource, delta Changelog) {
	for _, d := range delta {
		if d.Computed {
			// We need to copy the path (for console output compatibility) and remove the
			// last index if it's a slice.
			// We want a console output of format: struct.0.array.0: "foo" => "bar" (computed)
			// We want a json output of format: "message": "struct.0.array is a computed field"
			tmp := make([]string, len(d.Path))
			copy(tmp, d.Path)
			field, _ := a.getField(reflect.TypeOf(stateRes), tmp)
			if field.Type.Kind() == reflect.Slice {
				tmp = tmp[:len(tmp)-1]
			}
			path := strings.Join(tmp, ".")
			a.alerter.SendAlert(fmt.Sprintf("%s.%s", stateRes.TerraformType(), stateRes.TerraformId()),
				alerter.Alert{
					Message: fmt.Sprintf("%s is a computed field", path),
				})
		}
	}
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
			if field, ok := t.FieldByName(path[0]); ok && a.hasNestedFields(field.Type) {
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

package analyser

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/r3labs/diff/v2"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

type Change struct {
	diff.Change
	Computed   bool `json:"computed"`
	JsonString bool `json:"-"`
}

type Changelog []Change

type Difference struct {
	Res       resource.Resource
	Changelog Changelog
}

type Summary struct {
	TotalResources int `json:"total_resources"`
	TotalDrifted   int `json:"total_changed"`
	TotalUnmanaged int `json:"total_unmanaged"`
	TotalDeleted   int `json:"total_missing"`
	TotalManaged   int `json:"total_managed"`
}

type Analysis struct {
	unmanaged   []resource.Resource
	managed     []resource.Resource
	deleted     []resource.Resource
	differences []Difference
	summary     Summary
	alerts      alerter.Alerts
	Duration    time.Duration
	Date        time.Time
}

type serializableDifference struct {
	Res       resource.SerializableResource `json:"res"`
	Changelog Changelog                     `json:"changelog"`
}

type serializableAnalysis struct {
	Summary     Summary                                `json:"summary"`
	Managed     []resource.SerializableResource        `json:"managed"`
	Unmanaged   []resource.SerializableResource        `json:"unmanaged"`
	Deleted     []resource.SerializableResource        `json:"missing"`
	Differences []serializableDifference               `json:"differences"`
	Coverage    int                                    `json:"coverage"`
	Alerts      map[string][]alerter.SerializableAlert `json:"alerts"`
}

type GenDriftIgnoreOptions struct {
	ExcludeUnmanaged bool
	ExcludeDeleted   bool
	ExcludeDrifted   bool
	InputPath        string
}

func (a Analysis) MarshalJSON() ([]byte, error) {
	bla := serializableAnalysis{}
	for _, m := range a.managed {
		bla.Managed = append(bla.Managed, resource.SerializableResource{Resource: m})
	}
	for _, u := range a.unmanaged {
		bla.Unmanaged = append(bla.Unmanaged, resource.SerializableResource{Resource: u})
	}
	for _, d := range a.deleted {
		bla.Deleted = append(bla.Deleted, resource.SerializableResource{Resource: d})
	}
	for _, di := range a.differences {
		bla.Differences = append(bla.Differences, serializableDifference{
			Res:       resource.SerializableResource{Resource: di.Res},
			Changelog: di.Changelog,
		})
	}
	if len(a.alerts) > 0 {
		bla.Alerts = make(map[string][]alerter.SerializableAlert)
		for k, v := range a.alerts {
			for _, al := range v {
				bla.Alerts[k] = append(bla.Alerts[k], alerter.SerializableAlert{Alert: al})
			}
		}
	}
	bla.Summary = a.summary
	bla.Coverage = a.Coverage()

	return json.Marshal(bla)
}

func (a *Analysis) UnmarshalJSON(bytes []byte) error {
	bla := serializableAnalysis{}
	if err := json.Unmarshal(bytes, &bla); err != nil {
		return err
	}
	for _, u := range bla.Unmanaged {
		a.AddUnmanaged(&resource.SerializedResource{
			Id:   u.TerraformId(),
			Type: u.TerraformType(),
		})
	}
	for _, d := range bla.Deleted {
		a.AddDeleted(&resource.SerializedResource{
			Id:   d.TerraformId(),
			Type: d.TerraformType(),
		})
	}
	for _, m := range bla.Managed {
		a.AddManaged(&resource.SerializedResource{
			Id:   m.TerraformId(),
			Type: m.TerraformType(),
		})
	}
	for _, di := range bla.Differences {
		a.AddDifference(Difference{
			Res: &resource.SerializedResource{
				Id:   di.Res.TerraformId(),
				Type: di.Res.TerraformType(),
			},
			Changelog: di.Changelog,
		})
	}
	if len(bla.Alerts) > 0 {
		a.alerts = make(alerter.Alerts)
		for k, v := range bla.Alerts {
			for _, al := range v {
				a.alerts[k] = append(a.alerts[k], &alerter.SerializedAlert{
					Msg: al.Message(),
				})
			}
		}
	}
	return nil
}

func (a *Analysis) IsSync() bool {
	return a.summary.TotalDrifted == 0 && a.summary.TotalUnmanaged == 0 && a.summary.TotalDeleted == 0
}

func (a *Analysis) AddDeleted(resources ...resource.Resource) {
	a.deleted = append(a.deleted, resources...)
	a.summary.TotalResources += len(resources)
	a.summary.TotalDeleted += len(resources)
}

func (a *Analysis) AddUnmanaged(resources ...resource.Resource) {
	a.unmanaged = append(a.unmanaged, resources...)
	a.summary.TotalResources += len(resources)
	a.summary.TotalUnmanaged += len(resources)
}

func (a *Analysis) AddManaged(resources ...resource.Resource) {
	a.managed = append(a.managed, resources...)
	a.summary.TotalResources += len(resources)
	a.summary.TotalManaged += len(resources)
}

func (a *Analysis) AddDifference(diffs ...Difference) {
	a.differences = append(a.differences, diffs...)
	a.summary.TotalDrifted += len(diffs)
}

func (a *Analysis) SetAlerts(alerts alerter.Alerts) {
	a.alerts = alerts
}

func (a *Analysis) Coverage() int {
	if a.summary.TotalResources > 0 {
		return int((float32(a.summary.TotalManaged) / float32(a.summary.TotalResources)) * 100.0)
	}
	return 0
}

func (a *Analysis) Managed() []resource.Resource {
	return a.managed
}

func (a *Analysis) Unmanaged() []resource.Resource {
	return a.unmanaged
}

func (a *Analysis) Deleted() []resource.Resource {
	return a.deleted
}

func (a *Analysis) Differences() []Difference {
	return a.differences
}

func (a *Analysis) Summary() Summary {
	return a.summary
}

func (a *Analysis) Alerts() alerter.Alerts {
	return a.alerts
}

func (a *Analysis) SortResources() {
	a.unmanaged = resource.Sort(a.unmanaged)
	a.deleted = resource.Sort(a.deleted)
	a.differences = SortDifferences(a.differences)
}

func (a *Analysis) DriftIgnoreList(opts GenDriftIgnoreOptions) (int, string) {
	var list []string

	resourceCount := 0

	addResources := func(res ...resource.Resource) {
		for _, r := range res {
			list = append(list, fmt.Sprintf("%s.%s", r.TerraformType(), escapeKey(r.TerraformId())))
		}
		resourceCount += len(res)
	}
	addDifferences := func(diff ...Difference) {
		for _, d := range diff {
			addResources(d.Res)
		}
		resourceCount += len(diff)
	}

	if !opts.ExcludeUnmanaged && a.Summary().TotalUnmanaged > 0 {
		list = append(list, "# Resources not covered by IaC")
		addResources(a.Unmanaged()...)
	}
	if !opts.ExcludeDeleted && a.Summary().TotalDeleted > 0 {
		list = append(list, "# Missing resources")
		addResources(a.Deleted()...)
	}
	if !opts.ExcludeDrifted && a.Summary().TotalDrifted > 0 {
		list = append(list, "# Changed resources")
		addDifferences(a.Differences()...)
	}

	return resourceCount, strings.Join(list, "\n")
}

func SortDifferences(diffs []Difference) []Difference {
	sort.SliceStable(diffs, func(i, j int) bool {
		if diffs[i].Res.TerraformType() != diffs[j].Res.TerraformType() {
			return diffs[i].Res.TerraformType() < diffs[j].Res.TerraformType()
		}
		return diffs[i].Res.TerraformId() < diffs[j].Res.TerraformId()
	})

	for _, d := range diffs {
		SortChanges(d.Changelog)
	}

	return diffs
}

func SortChanges(changes []Change) []Change {
	sort.SliceStable(changes, func(i, j int) bool {
		return strings.Join(changes[i].Path, ".") < strings.Join(changes[j].Path, ".")
	})
	return changes
}

func escapeKey(line string) string {
	line = strings.ReplaceAll(line, `\`, `\\`)
	line = strings.ReplaceAll(line, `.`, `\.`)

	return line
}

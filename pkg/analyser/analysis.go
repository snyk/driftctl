package analyser

import (
	"encoding/json"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/r3labs/diff/v2"
)

type Change struct {
	diff.Change
	Computed bool `json:"computed"`
}

type Changelog []Change

type Difference struct {
	Res       resource.Resource
	Changelog Changelog
}

type Summary struct {
	TotalResources int `json:"total_resources"`
	TotalDrifted   int `json:"total_drifted"`
	TotalUnmanaged int `json:"total_unmanaged"`
	TotalDeleted   int `json:"total_deleted"`
	TotalManaged   int `json:"total_managed"`
}

type Analysis struct {
	unmanaged   []resource.Resource
	managed     []resource.Resource
	deleted     []resource.Resource
	differences []Difference
	summary     Summary
	alerts      alerter.Alerts
}

type serializableDifference struct {
	Res       resource.SerializableResource `json:"res"`
	Changelog Changelog                     `json:"changelog"`
}

type serializableAnalysis struct {
	Summary     Summary                         `json:"summary"`
	Managed     []resource.SerializableResource `json:"managed"`
	Unmanaged   []resource.SerializableResource `json:"unmanaged"`
	Deleted     []resource.SerializableResource `json:"deleted"`
	Differences []serializableDifference        `json:"differences"`
	Coverage    int                             `json:"coverage"`
	Alerts      alerter.Alerts                  `json:"alerts"`
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
	bla.Summary = a.summary
	bla.Coverage = a.Coverage()
	bla.Alerts = a.alerts

	return json.Marshal(bla)
}

func (a *Analysis) UnmarshalJSON(bytes []byte) error {
	bla := serializableAnalysis{}
	if err := json.Unmarshal(bytes, &bla); err != nil {
		return err
	}
	for _, u := range bla.Unmanaged {
		a.AddUnmanaged(resource.SerializedResource{
			Id:   u.TerraformId(),
			Type: u.TerraformType(),
		})
	}
	for _, d := range bla.Deleted {
		a.AddDeleted(resource.SerializedResource{
			Id:   d.TerraformId(),
			Type: d.TerraformType(),
		})
	}
	for _, m := range bla.Managed {
		a.AddManaged(resource.SerializedResource{
			Id:   m.TerraformId(),
			Type: m.TerraformType(),
		})
	}
	for _, di := range bla.Differences {
		a.AddDifference(Difference{
			Res: resource.SerializedResource{
				Id:   di.Res.TerraformId(),
				Type: di.Res.TerraformType(),
			},
			Changelog: di.Changelog,
		})
	}
	a.SetAlerts(bla.Alerts)
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

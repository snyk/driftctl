package analyser

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/snyk/driftctl/enumeration/alerter"

	"github.com/snyk/driftctl/enumeration/resource"
)

type Summary struct {
	TotalResources      int  `json:"total_resources"`
	TotalUnmanaged      int  `json:"total_unmanaged"`
	TotalDeleted        int  `json:"total_missing"`
	TotalManaged        int  `json:"total_managed"`
	TotalIaCSourceCount uint `json:"total_iac_source_count"`
}

type Analysis struct {
	unmanaged       []*resource.Resource
	managed         []*resource.Resource
	deleted         []*resource.Resource
	summary         Summary
	alerts          alerter.Alerts
	Duration        time.Duration
	Date            time.Time
	ProviderName    string
	ProviderVersion string
}

type serializableAnalysis struct {
	Summary         Summary                                `json:"summary"`
	Managed         []resource.SerializableResource        `json:"managed"`
	Unmanaged       []resource.SerializableResource        `json:"unmanaged"`
	Deleted         []resource.SerializableResource        `json:"missing"`
	Coverage        int                                    `json:"coverage"`
	Alerts          map[string][]alerter.SerializableAlert `json:"alerts"`
	ProviderName    string                                 `json:"provider_name"`
	ProviderVersion string                                 `json:"provider_version"`
	ScanDuration    uint                                   `json:"scan_duration,omitempty"`
	Date            time.Time                              `json:"date"`
}

type GenDriftIgnoreOptions struct {
	ExcludeUnmanaged bool
	ExcludeDeleted   bool
	ExcludeDrifted   bool
	InputPath        string
	OutputPath       string
}

func NewAnalysis() *Analysis {
	return &Analysis{}
}

func (a Analysis) MarshalJSON() ([]byte, error) {
	bla := serializableAnalysis{}
	for _, m := range a.managed {
		bla.Managed = append(bla.Managed, *resource.NewSerializableResource(m))
	}
	for _, u := range a.unmanaged {
		bla.Unmanaged = append(bla.Unmanaged, *resource.NewSerializableResource(u))
	}
	for _, d := range a.deleted {
		bla.Deleted = append(bla.Deleted, *resource.NewSerializableResource(d))
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
	bla.ProviderName = a.ProviderName
	bla.ProviderVersion = a.ProviderVersion
	bla.ScanDuration = uint(a.Duration.Seconds())
	bla.Date = a.Date

	return json.Marshal(bla)
}

func (a *Analysis) UnmarshalJSON(bytes []byte) error {
	bla := serializableAnalysis{}
	if err := json.Unmarshal(bytes, &bla); err != nil {
		return err
	}
	for _, u := range bla.Unmanaged {
		a.AddUnmanaged(&resource.Resource{
			Id:   u.Id,
			Type: u.Type,
		})
	}
	for _, d := range bla.Deleted {
		a.AddDeleted(&resource.Resource{
			Id:   d.Id,
			Type: d.Type,
		})
	}
	for _, m := range bla.Managed {
		res := &resource.Resource{
			Id:   m.Id,
			Type: m.Type,
		}
		if m.Source != nil {
			// We loose the source type in the serialization process, for now everything is serialized back to a
			// TerraformStateSource.
			// TODO: Add a discriminator field to be able to serialize back to the right type
			// when we'll introduce a new source type
			res.Source = &resource.TerraformStateSource{
				State:  m.Source.S,
				Module: m.Source.Ns,
				Name:   m.Source.Name,
			}
		}
		a.AddManaged(res)
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
	a.ProviderName = bla.ProviderName
	a.ProviderVersion = bla.ProviderVersion
	a.SetIaCSourceCount(bla.Summary.TotalIaCSourceCount)
	a.Duration = time.Duration(bla.ScanDuration) * time.Second
	a.Date = bla.Date
	return nil
}

func (a *Analysis) IsSync() bool {
	return a.summary.TotalUnmanaged == 0 && a.summary.TotalDeleted == 0
}

func (a *Analysis) AddDeleted(resources ...*resource.Resource) {
	a.deleted = append(a.deleted, resources...)
	a.summary.TotalResources += len(resources)
	a.summary.TotalDeleted += len(resources)
}

func (a *Analysis) AddUnmanaged(resources ...*resource.Resource) {
	a.unmanaged = append(a.unmanaged, resources...)
	a.summary.TotalResources += len(resources)
	a.summary.TotalUnmanaged += len(resources)
}

func (a *Analysis) AddManaged(resources ...*resource.Resource) {
	a.managed = append(a.managed, resources...)
	a.summary.TotalResources += len(resources)
	a.summary.TotalManaged += len(resources)
}

func (a *Analysis) SetAlerts(alerts alerter.Alerts) {
	a.alerts = alerts
}

func (a *Analysis) SetIaCSourceCount(i uint) {
	a.summary.TotalIaCSourceCount = i
}

func (a *Analysis) Coverage() int {
	if a.summary.TotalResources > 0 {
		return int((float32(a.summary.TotalManaged) / float32(a.summary.TotalResources)) * 100.0)
	}
	return 0
}

func (a *Analysis) Managed() []*resource.Resource {
	return a.managed
}

func (a *Analysis) Unmanaged() []*resource.Resource {
	return a.unmanaged
}

func (a *Analysis) Deleted() []*resource.Resource {
	return a.deleted
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
}

func (a *Analysis) DriftIgnoreList(opts GenDriftIgnoreOptions) (int, string) {
	var list []string

	resourceCount := 0

	addResources := func(res ...*resource.Resource) {
		for _, r := range res {
			list = append(list, fmt.Sprintf("%s.%s", r.ResourceType(), escapeKey(r.ResourceId())))
		}
		resourceCount += len(res)
	}

	if !opts.ExcludeUnmanaged && a.Summary().TotalUnmanaged > 0 {
		list = append(list, "# Resources not covered by IaC")
		addResources(a.Unmanaged()...)
	}
	if !opts.ExcludeDeleted && a.Summary().TotalDeleted > 0 {
		list = append(list, "# Missing resources")
		addResources(a.Deleted()...)
	}

	return resourceCount, strings.Join(list, "\n")
}

func escapeKey(line string) string {
	line = strings.ReplaceAll(line, `\`, `\\`)
	line = strings.ReplaceAll(line, `.`, `\.`)

	return line
}

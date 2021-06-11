package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const FormatVersion = "0.1"
const PlanOutputType = "plan"
const PlanOutputExample = "plan://PATH/TO/FILE.json"

type plan struct {
	FormatVersion   string        `json:"format_version,omitempty"`
	PlannedValues   plannedValues `json:"planned_values,omitempty"`
	ResourceChanges []rscChange   `json:"resource_changes,omitempty"`
}

type plannedValues struct {
	RootModule module `json:"root_module,omitempty"`
}

type rscChange struct {
	Address string `json:"address,omitempty"`
	Type    string `json:"type,omitempty"`
	Name    string `json:"name,omitempty"`
	Change  change `json:"change,omitempty"`
}

type change struct {
	Actions []string               `json:"actions,omitempty"`
	Before  map[string]interface{} `json:"before,omitempty"`
	After   map[string]interface{} `json:"after,omitempty"`
}

type module struct {
	Resources []rsc `json:"resources,omitempty"`
}

type rsc struct {
	Address         string                 `json:"address,omitempty"`
	Type            string                 `json:"type,omitempty"`
	Name            string                 `json:"name,omitempty"`
	AttributeValues map[string]interface{} `json:"values,omitempty"`
}

type Plan struct {
	path string
}

func NewPlan(path string) *Plan {
	return &Plan{path}
}

func (c *Plan) Write(analysis *analyser.Analysis) error {
	file := os.Stdout
	if !isStdOut(c.path) {
		f, err := os.OpenFile(c.path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		file = f
	}
	output := plan{FormatVersion: FormatVersion}
	output.PlannedValues.RootModule = addPlannedValues(analysis)
	output.ResourceChanges = addResourceChanges(analysis)
	jsonPlan, err := json.MarshalIndent(output, "", "\t")
	if err != nil {
		return err
	}
	if _, err := file.Write(jsonPlan); err != nil {
		return err
	}
	return nil
}

func addPlannedValues(analysis *analyser.Analysis) module {
	managedRsc := listRsc(analysis.Managed())
	unmanagedRsc := listRsc(analysis.Unmanaged())
	return module{
		Resources: append(managedRsc, unmanagedRsc...),
	}
}

func listRsc(resources []resource.Resource) []rsc {
	var ret []rsc
	for _, res := range resources {
		r := rsc{
			Address:         fmt.Sprintf("%s.%s", res.TerraformType(), res.TerraformId()),
			Type:            res.TerraformType(),
			Name:            res.TerraformId(),
			AttributeValues: *res.Attributes(),
		}
		ret = append(ret, r)
	}
	return ret
}

func addResourceChanges(analysis *analyser.Analysis) []rscChange {
	managedRsc := listRscChange(analysis.Managed(), "no-op")
	unmanagedRsc := listRscChange(analysis.Unmanaged(), "create")
	return append(managedRsc, unmanagedRsc...)
}

func listRscChange(resources []resource.Resource, action string) []rscChange {
	var ret []rscChange
	for _, res := range resources {
		r := rscChange{
			Address: fmt.Sprintf("%s.%s", res.TerraformType(), res.TerraformId()),
			Type:    res.TerraformType(),
			Name:    res.TerraformId(),
			Change: change{
				Actions: []string{action},
				After:   *res.Attributes(),
			},
		}
		if action == "no-op" {
			r.Change.Before = *res.Attributes()
		}
		ret = append(ret, r)

	}
	return ret
}

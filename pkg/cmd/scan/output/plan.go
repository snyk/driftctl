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
	output.PlannedValues.RootModule = addPlannedValues(analysis.Unmanaged())
	output.ResourceChanges = addResourceChanges(analysis.Unmanaged())
	jsonPlan, err := json.MarshalIndent(output, "", "\t")
	if err != nil {
		return err
	}
	if _, err := file.Write(jsonPlan); err != nil {
		return err
	}
	return nil
}

func addPlannedValues(resources []resource.Resource) module {
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
	return module{
		Resources: ret,
	}
}

func addResourceChanges(resources []resource.Resource) []rscChange {
	var ret []rscChange
	for _, res := range resources {
		r := rscChange{
			Address: fmt.Sprintf("%s.%s", res.TerraformType(), res.TerraformId()),
			Type:    res.TerraformType(),
			Name:    res.TerraformId(),
			Change: change{
				Actions: []string{"create"},
				After:   *res.Attributes(),
			},
		}
		ret = append(ret, r)

	}
	return ret
}

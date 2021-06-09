package output

import (
	"encoding/json"
	"os"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/github"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/command/jsonplan"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/hashicorp/terraform/plans"
	tf "github.com/hashicorp/terraform/terraform"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

const PlanOutputType = "plan"
const PlanOutputExample = "plan://PATH/TO/FILE.json"

type Plan struct {
	path   string
	remote string
}

func NewPlan(path, remote string) *Plan {
	switch remote {
	case aws.RemoteAWSTerraform:
		return &Plan{path: path, remote: "aws"}
	case github.RemoteGithubTerraform:
		return &Plan{path: path, remote: "github"}
	default:
		return &Plan{path: path}
	}
}

func (c *Plan) Write(analysis *analyser.Analysis, providerLibrary *terraform.ProviderLibrary) error {
	file := os.Stdout
	if !isStdOut(c.path) {
		f, err := os.OpenFile(c.path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		file = f
	}

	schemas := &tf.Schemas{
		Providers: map[addrs.Provider]*tf.ProviderSchema{},
	}

	for k, p := range providerLibrary.Providers() {
		resp := p.TerraformProviderSchema()
		s := &tf.ProviderSchema{
			Provider:                   resp.Provider.Block,
			ResourceTypes:              make(map[string]*configschema.Block),
			ResourceTypeSchemaVersions: make(map[string]uint64),
		}
		for t, r := range resp.ResourceTypes {
			s.ResourceTypes[t] = r.Block
			s.ResourceTypeSchemaVersions[t] = uint64(r.Version)
		}
		schemas.Providers[addrs.NewDefaultProvider(k)] = s
	}

	var resources []*plans.ResourceInstanceChangeSrc
	for _, res := range analysis.Unmanaged() {
		attrs, err := json.Marshal(res.Attributes())
		if err != nil {
			return err
		}
		schema, _ := schemas.ResourceTypeConfig(addrs.NewDefaultProvider(c.remote), addrs.ManagedResourceMode, res.TerraformType())
		ty := schema.ImpliedType()
		ctyVal, err := ctyjson.Unmarshal(attrs, ty)
		if err != nil {
			return err
		}
		val, err := plans.NewDynamicValue(ctyVal, ty)
		if err != nil {
			return err
		}
		resources = append(resources, &plans.ResourceInstanceChangeSrc{
			Addr: addrs.Resource{
				Mode: addrs.ManagedResourceMode,
				Type: res.TerraformType(),
				Name: res.TerraformId(),
			}.Instance(addrs.IntKey(0)).Absolute(addrs.RootModuleInstance),
			ProviderAddr: addrs.AbsProviderConfig{
				Module:   addrs.RootModule,
				Provider: addrs.NewDefaultProvider(c.remote),
			},
			ChangeSrc: plans.ChangeSrc{
				Action: plans.NoOp,
				After:  val,
			},
		})
	}

	plan := &plans.Plan{
		Changes: &plans.Changes{
			Resources: resources,
		},
	}

	jsonPlan, err := jsonplan.Marshal(configs.NewEmptyConfig(), plan, nil, schemas)
	if err != nil {
		return err
	}
	if _, err := file.Write(jsonPlan); err != nil {
		return err
	}
	return nil
}

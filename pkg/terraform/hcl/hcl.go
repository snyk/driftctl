package hcl

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type MainBodyBlock struct {
	Terraform TerraformBlock `hcl:"terraform,block"`
	Remain    hcl.Body       `hcl:",remain"`
}

type TerraformBlock struct {
	Backend BackendBlock `hcl:"backend,block"`
	Remain  hcl.Body     `hcl:",remain"`
}

func ParseTerraformFromHCL(filename string) (*TerraformBlock, error) {
	var body MainBodyBlock

	body.Terraform.Backend.workspace = getCurrentWorkspaceName(path.Dir(filename))

	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, diags
	}

	diags = gohcl.DecodeBody(f.Body, nil, &body)
	if diags.HasErrors() {
		return nil, diags
	}

	return &body.Terraform, nil
}

func getCurrentWorkspaceName(cwd string) string {
	env := "default" // See https://github.com/hashicorp/terraform/blob/main/internal/backend/backend.go#L33

	data, err := ioutil.ReadFile(path.Join(cwd, ".terraform/environment"))
	if err != nil {
		return env
	}
	return strings.Trim(string(data), "\n")
}

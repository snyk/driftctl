package hcl

import (
	"os"
	"path"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

const DefaultStateName = "default"

type MainBodyBlock struct {
	Terraform TerraformBlock `hcl:"terraform,block"`
	Remain    hcl.Body       `hcl:",remain"`
}

type TerraformBlock struct {
	Backend *BackendBlock `hcl:"backend,block"`
	Cloud   *CloudBlock   `hcl:"cloud,block"`
	Remain  hcl.Body      `hcl:",remain"`
}

func ParseTerraformFromHCL(filename string) (*TerraformBlock, error) {
	var body MainBodyBlock

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

func GetCurrentWorkspaceName(cwd string) string {
	name := DefaultStateName // See https://github.com/hashicorp/terraform/blob/main/internal/backend/backend.go#L33

	data, err := os.ReadFile(path.Join(cwd, ".terraform/environment"))
	if err != nil {
		return name
	}
	if v := strings.Trim(string(data), "\n"); v != "" {
		name = v
	}
	return name
}

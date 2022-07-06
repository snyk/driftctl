package hcl

import (
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type MainBodyBlock struct {
	Terraform TerraformBlock `hcl:"terraform,block"`
}

type TerraformBlock struct {
	Backend BackendBlock `hcl:"backend,block"`
}

func ParseTerraformFromHCL(filename string) (*TerraformBlock, error) {
	var v MainBodyBlock

	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, diags
	}

	diags = gohcl.DecodeBody(f.Body, nil, &v)
	if diags.HasErrors() {
		return nil, diags
	}

	return &v.Terraform, nil
}

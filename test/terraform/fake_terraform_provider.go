package terraform

import (
	gojson "encoding/json"
	"fmt"
	"sort"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/cloudskiff/driftctl/test/mocks"
	"github.com/hashicorp/terraform/providers"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
)

type FakeTerraformProvider struct {
	realProvider terraform.TerraformProvider
	shouldUpdate bool
	response     string
}

func NewFakeTerraformProvider(realProvider terraform.TerraformProvider) *FakeTerraformProvider {
	return &FakeTerraformProvider{realProvider: realProvider}
}

func (p *FakeTerraformProvider) ShouldUpdate() {
	p.shouldUpdate = true
}

func (p *FakeTerraformProvider) Schema() map[string]providers.Schema {
	return p.readSchema()
}

func (p *FakeTerraformProvider) WithResponse(response string) *FakeTerraformProvider {
	p.response = response
	return p
}

func (p *FakeTerraformProvider) ReadResource(args terraform.ReadResourceArgs) (*cty.Value, error) {
	if p.response == "" {
		return nil, errors.New("WithResponse should be called before ReadResource to specify a directory to fetch fake response")
	}
	if p.shouldUpdate {
		readResource, err := p.realProvider.ReadResource(args)
		p.writeResource(args, readResource, err)
		return readResource, err
	}

	return p.readResource(args)
}

func (p *FakeTerraformProvider) readSchema() map[string]providers.Schema {
	content, err := test.ReadTestFile(fmt.Sprintf("../schemas/%s/%s/schema.json", p.realProvider.Name(), p.realProvider.Version()))
	if err != nil {
		panic(err)
	}
	var schema map[string]providers.Schema
	if err := gojson.Unmarshal(content, &schema); err != nil {
		panic(err)
	}
	return schema
}

func (p *FakeTerraformProvider) writeResource(args terraform.ReadResourceArgs, readResource *cty.Value, err error) {
	var readRes = mocks.ReadResource{
		Value: readResource,
		Err:   err,
	}

	marshalled, err := gojson.Marshal(&readRes)
	if err != nil {
		panic(err)
	}
	fileName := p.getFileName(args)
	goldenfile.WriteFile(p.response, marshalled, fileName)
}

func (p *FakeTerraformProvider) readResource(args terraform.ReadResourceArgs) (*cty.Value, error) {
	fileName := p.getFileName(args)
	content := goldenfile.ReadFile(p.response, fileName)
	var readRes mocks.ReadResource
	if err := gojson.Unmarshal(content, &readRes); err != nil {
		panic(err)
	}
	return readRes.Value, readRes.Err
}

func (p *FakeTerraformProvider) getFileName(args terraform.ReadResourceArgs) string {
	suffix := ""
	keys := make([]string, 0, len(args.Attributes))
	for k := range args.Attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		suffix = fmt.Sprintf("%s-%s", suffix, args.Attributes[k])
	}
	fileName := fmt.Sprintf("%s-%s%s.res.golden.json", args.Ty, args.ID, suffix)
	return fileName
}

func (p *FakeTerraformProvider) Cleanup() {}

func (p *FakeTerraformProvider) Name() string {
	return p.realProvider.Name()
}

func (p *FakeTerraformProvider) Version() string {
	return p.realProvider.Version()
}

package terraform

import (
	"crypto/sha1"
	gojson "encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/cloudskiff/driftctl/test/mocks"
	"github.com/cloudskiff/driftctl/test/schemas"
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
	schema, err := schemas.ReadTestSchema(p.realProvider.Name(), p.realProvider.Version())
	if err != nil {
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

	// ext4 and many other filesystems has a maximum filename length of 255 bytes
	// See https://en.wikipedia.org/wiki/Comparison_of_file_systems#Limits
	// Solution: we create a SHA1 hash of the filename so the length stay constant
	// We should do that no matter the length, but it requires to regenerate every single file
	// TODO: Use SHA1 filenames for all resource golden files
	resourceUID := fmt.Sprintf("%s-%s%s", args.Ty, args.ID, suffix)
	if len(resourceUID) > 239 {
		h := sha1.New()
		_, _ = io.WriteString(h, resourceUID)
		resourceUID = fmt.Sprintf("%x", h.Sum(nil))
	}

	fileName := fmt.Sprintf("%s.res.golden.json", resourceUID)
	return fileName
}

func (p *FakeTerraformProvider) Cleanup() {}

func (p *FakeTerraformProvider) Name() string {
	return p.realProvider.Name()
}

func (p *FakeTerraformProvider) Version() string {
	return p.realProvider.Version()
}

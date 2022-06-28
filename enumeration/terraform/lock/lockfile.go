package lock

import (
	"strings"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type ProviderBlock struct {
	Address     string   `hcl:"address,label"`
	Version     string   `hcl:"version,attr"`
	Constraints string   `hcl:"constraints,optional"`
	Hashes      []string `hcl:"hashes,optional"`
}

// ProviderAddress encapsulates a single provider type. In the future this will be
// extended to include additional fields including Namespace and SourceHost
type ProviderAddress struct {
	Type      string
	Namespace string
	Hostname  string
}

func (p *ProviderAddress) String() string {
	return strings.Join([]string{p.Hostname, p.Namespace, p.Type}, "/")
}

type Lockfile struct {
	Providers []ProviderBlock `hcl:"provider,block"`
}

func (l *Lockfile) GetProviderByAddress(addr *ProviderAddress) *ProviderBlock {
	for _, p := range l.Providers {
		if p.Address == addr.String() {
			return &p
		}
	}
	return nil
}

func ReadLocksFromFile(filename string) (*Lockfile, error) {
	var lock Lockfile

	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return &lock, diags
	}

	diags = gohcl.DecodeBody(f.Body, nil, &lock)

	if diags.HasErrors() {
		return &lock, diags
	}

	return &lock, nil
}

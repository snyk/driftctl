package filter

import (
	"github.com/jmespath/go-jmespath"
	"github.com/snyk/driftctl/enumeration/resource"
)

type CompositeFilter struct {
	driftIgnore Filter
	jmesPath    *jmespath.JMESPath
}

func NewCompositeFilter(driftIgnore Filter, jmesPath *jmespath.JMESPath) *CompositeFilter {
	return &CompositeFilter{
		driftIgnore: driftIgnore,
		jmesPath:    jmesPath,
	}
}

func (f *CompositeFilter) IsTypeIgnored(ty resource.ResourceType) bool {
	if f.driftIgnore.IsTypeIgnored(ty) {
		return true
	}

	if f.jmesPath == nil {
		return false
	}

	filtrable := filtrableResource{
		Type: string(ty),
	}

	result, err := f.jmesPath.Search([]filtrableResource{filtrable})
	if err != nil {
		return false
	}

	results, ok := result.([]interface{})
	return !ok || len(results) == 0
}

func (f *CompositeFilter) IsResourceIgnored(res *resource.Resource) bool {
	return f.driftIgnore.IsResourceIgnored(res)
}

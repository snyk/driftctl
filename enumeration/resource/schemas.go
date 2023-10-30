package resource

import (
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform/configs/configschema"
)

type AttributeSchema struct {
	ConfigSchema configschema.Attribute
	JsonString   bool
}

type Flags uint32

func (f Flags) HasFlag(flag Flags) bool {
	return f&flag != 0
}

func (f *Flags) AddFlag(flag Flags) {
	*f |= flag
}

type Schema struct {
	ProviderVersion             *version.Version
	Flags                       Flags
	SchemaVersion               int64
	Attributes                  map[string]AttributeSchema
	NormalizeFunc               func(res *Resource)
	HumanReadableAttributesFunc func(res *Resource) map[string]string
	DiscriminantFunc            func(*Resource, *Resource) bool
}

func (s *Schema) IsComputedField(path []string) bool {
	metadata, exist := s.Attributes[strings.Join(path, ".")]
	if !exist {
		return false
	}
	return metadata.ConfigSchema.Computed
}

func (s *Schema) IsJsonStringField(path []string) bool {
	metadata, exist := s.Attributes[strings.Join(path, ".")]
	if !exist {
		return false
	}
	return metadata.JsonString
}

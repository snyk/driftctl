package resource

import (
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/hashicorp/terraform/providers"
	"github.com/sirupsen/logrus"
)

type AttributeSchema struct {
	ConfigSchema configschema.Attribute
	JsonString   bool
}

type Flags uint32

const (
	FlagDeepMode Flags = 1 << iota
)

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

type SchemaRepositoryInterface interface {
	GetSchema(resourceType string) (*Schema, bool)
	SetFlags(typ string, flags ...Flags)
	UpdateSchema(typ string, schemasMutators map[string]func(attributeSchema *AttributeSchema))
	SetNormalizeFunc(typ string, normalizeFunc func(res *Resource))
	SetHumanReadableAttributesFunc(typ string, humanReadableAttributesFunc func(res *Resource) map[string]string)
	SetDiscriminantFunc(string, func(*Resource, *Resource) bool)
}

type SchemaRepository struct {
	schemas         map[string]*Schema
	ProviderName    string
	ProviderVersion *version.Version
}

func NewSchemaRepository() *SchemaRepository {
	return &SchemaRepository{
		schemas: make(map[string]*Schema),
	}
}

func (r *SchemaRepository) GetSchema(resourceType string) (*Schema, bool) {
	schema, exist := r.schemas[resourceType]
	return schema, exist
}

func (r *SchemaRepository) fetchNestedBlocks(root string, metadata map[string]AttributeSchema, block map[string]*configschema.NestedBlock) {
	for s, nestedBlock := range block {
		path := s
		if root != "" {
			path = strings.Join([]string{root, s}, ".")
		}
		for s2, attr := range nestedBlock.Attributes {
			nestedPath := strings.Join([]string{path, s2}, ".")
			metadata[nestedPath] = AttributeSchema{
				ConfigSchema: *attr,
			}
		}
		r.fetchNestedBlocks(path, metadata, nestedBlock.BlockTypes)
	}
}

func (r *SchemaRepository) Init(providerName, providerVersion string, schema map[string]providers.Schema) error {
	v, err := version.NewVersion(providerVersion)
	if err != nil {
		return err
	}
	r.ProviderVersion = v
	r.ProviderName = providerName
	for typ, sch := range schema {
		attributeMetas := map[string]AttributeSchema{}
		for s, attribute := range sch.Block.Attributes {
			attributeMetas[s] = AttributeSchema{
				ConfigSchema: *attribute,
			}
		}

		r.fetchNestedBlocks("", attributeMetas, sch.Block.BlockTypes)

		r.schemas[typ] = &Schema{
			ProviderVersion: r.ProviderVersion,
			SchemaVersion:   sch.Version,
			Attributes:      attributeMetas,
		}
	}
	return nil
}

func (r SchemaRepository) SetFlags(typ string, flags ...Flags) {
	metadata, exist := r.GetSchema(typ)
	if !exist {
		logrus.WithFields(logrus.Fields{"type": typ}).Warning("Unable to set flags, no schema found")
		return
	}
	for _, flag := range flags {
		metadata.Flags.AddFlag(flag)
	}
}

func (r *SchemaRepository) UpdateSchema(typ string, schemasMutators map[string]func(attributeSchema *AttributeSchema)) {
	for s, f := range schemasMutators {
		metadata, exist := r.GetSchema(typ)
		if !exist {
			logrus.WithFields(logrus.Fields{"type": typ}).Warning("Unable to set metadata, no schema found")
			return
		}
		m := (*metadata).Attributes[s]
		f(&m)
		(*metadata).Attributes[s] = m
	}
}

func (r *SchemaRepository) SetNormalizeFunc(typ string, normalizeFunc func(res *Resource)) {
	metadata, exist := r.GetSchema(typ)
	if !exist {
		logrus.WithFields(logrus.Fields{"type": typ}).Warning("Unable to set normalize func, no schema found")
		return
	}
	(*metadata).NormalizeFunc = normalizeFunc
}

func (r *SchemaRepository) SetHumanReadableAttributesFunc(typ string, humanReadableAttributesFunc func(res *Resource) map[string]string) {
	metadata, exist := r.GetSchema(typ)
	if !exist {
		logrus.WithFields(logrus.Fields{"type": typ}).Warning("Unable to add human readable attributes, no schema found")
		return
	}
	(*metadata).HumanReadableAttributesFunc = humanReadableAttributesFunc
}

func (r *SchemaRepository) SetDiscriminantFunc(typ string, fn func(self, res *Resource) bool) {
	metadata, exist := r.GetSchema(typ)
	if !exist {
		logrus.WithFields(logrus.Fields{"type": typ}).Warning("Unable to set discriminant function, no schema found")
		return
	}
	(*metadata).DiscriminantFunc = fn
}

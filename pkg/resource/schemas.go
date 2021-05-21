package resource

import (
	"strings"

	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/hashicorp/terraform/providers"
	"github.com/sirupsen/logrus"
)

type AttributeSchema struct {
	ConfigSchema configschema.Attribute
	JsonString   bool
}

type Schema struct {
	Attributes                  map[string]AttributeSchema
	NormalizeFunc               func(res *AbstractResource)
	HumanReadableAttributesFunc func(res *AbstractResource) map[string]string
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
	UpdateSchema(typ string, schemasMutators map[string]func(attributeSchema *AttributeSchema))
	SetNormalizeFunc(typ string, normalizeFunc func(res *AbstractResource))
	SetHumanReadableAttributesFunc(typ string, humanReadableAttributesFunc func(res *AbstractResource) map[string]string)
}

type SchemaRepository struct {
	schemas map[string]*Schema
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

func (r *SchemaRepository) Init(schema map[string]providers.Schema) {
	for typ, sch := range schema {
		attributeMetas := map[string]AttributeSchema{}
		for s, attribute := range sch.Block.Attributes {
			attributeMetas[s] = AttributeSchema{
				ConfigSchema: *attribute,
			}
		}

		r.fetchNestedBlocks("", attributeMetas, sch.Block.BlockTypes)

		r.schemas[typ] = &Schema{
			Attributes: attributeMetas,
		}
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

func (r *SchemaRepository) SetNormalizeFunc(typ string, normalizeFunc func(res *AbstractResource)) {
	metadata, exist := r.GetSchema(typ)
	if !exist {
		logrus.WithFields(logrus.Fields{"type": typ}).Warning("Unable to set normalize func, no schema found")
		return
	}
	(*metadata).NormalizeFunc = normalizeFunc
}

func (r *SchemaRepository) SetHumanReadableAttributesFunc(typ string, humanReadableAttributesFunc func(res *AbstractResource) map[string]string) {
	metadata, exist := r.GetSchema(typ)
	if !exist {
		logrus.WithFields(logrus.Fields{"type": typ}).Warning("Unable to add human readable attributes, no schema found")
		return
	}
	(*metadata).HumanReadableAttributesFunc = humanReadableAttributesFunc
}

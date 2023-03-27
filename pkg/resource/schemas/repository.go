package schemas

import (
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/hashicorp/terraform/providers"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/snyk/driftctl/pkg/resource/azurerm"
	"github.com/snyk/driftctl/pkg/resource/github"
	"github.com/snyk/driftctl/pkg/resource/google"
	"github.com/snyk/driftctl/pkg/resource/scaleway"
)

type SchemaRepository struct {
	schemas map[string]*resource.Schema
}

func NewSchemaRepository() *SchemaRepository {
	return &SchemaRepository{
		schemas: make(map[string]*resource.Schema),
	}
}

func (r *SchemaRepository) GetSchema(resourceType string) (*resource.Schema, bool) {
	schema, exist := r.schemas[resourceType]
	return schema, exist
}

func (r *SchemaRepository) fetchNestedBlocks(root string, metadata map[string]resource.AttributeSchema, block map[string]*configschema.NestedBlock) {
	for s, nestedBlock := range block {
		path := s
		if root != "" {
			path = strings.Join([]string{root, s}, ".")
		}
		for s2, attr := range nestedBlock.Attributes {
			nestedPath := strings.Join([]string{path, s2}, ".")
			metadata[nestedPath] = resource.AttributeSchema{
				ConfigSchema: *attr,
			}
		}
		r.fetchNestedBlocks(path, metadata, nestedBlock.BlockTypes)
	}
}

func (r *SchemaRepository) Init(providerName, providerVersion string, schema map[string]providers.Schema) error {

	if providerVersion == "" {
		switch providerName {
		case "aws":
			providerVersion = "3.19.0"
		case "github":
			providerVersion = "4.4.0"
		case "google":
			providerVersion = "3.78.0"
		case "azurerm":
			providerVersion = "2.71.0"
		case "scaleway":
			providerVersion = "2.14.1"
		default:
			return errors.Errorf("unsupported remote '%s'", providerName)
		}
	}

	v, err := version.NewVersion(providerVersion)
	if err != nil {
		return err
	}
	for typ, sch := range schema {
		attributeMetas := map[string]resource.AttributeSchema{}
		for s, attribute := range sch.Block.Attributes {
			attributeMetas[s] = resource.AttributeSchema{
				ConfigSchema: *attribute,
			}
		}

		r.fetchNestedBlocks("", attributeMetas, sch.Block.BlockTypes)

		r.schemas[typ] = &resource.Schema{
			ProviderVersion: v,
			SchemaVersion:   sch.Version,
			Attributes:      attributeMetas,
		}
	}
	switch providerName {
	case "aws":
		aws.InitResourcesMetadata(r)
	case "github":
		github.InitResourcesMetadata(r)
	case "google":
		google.InitResourcesMetadata(r)
	case "azurerm":
		azurerm.InitResourcesMetadata(r)
	case "scaleway":
		scaleway.InitResourcesMetadata(r)
	default:
		return errors.Errorf("unsupported remote '%s'", providerName)
	}
	return nil
}

func (r SchemaRepository) SetFlags(typ string, flags ...resource.Flags) {
	metadata, exist := r.GetSchema(typ)
	if !exist {
		logrus.WithFields(logrus.Fields{"type": typ}).Warning("Unable to set flags, no schema found")
		return
	}
	for _, flag := range flags {
		metadata.Flags.AddFlag(flag)
	}
}

func (r *SchemaRepository) UpdateSchema(typ string, schemasMutators map[string]func(attributeSchema *resource.AttributeSchema)) {
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

func (r *SchemaRepository) SetNormalizeFunc(typ string, normalizeFunc func(res *resource.Resource)) {
	metadata, exist := r.GetSchema(typ)
	if !exist {
		logrus.WithFields(logrus.Fields{"type": typ}).Warning("Unable to set normalize func, no schema found")
		return
	}
	(*metadata).NormalizeFunc = normalizeFunc
}

func (r *SchemaRepository) SetHumanReadableAttributesFunc(typ string, humanReadableAttributesFunc func(res *resource.Resource) map[string]string) {
	metadata, exist := r.GetSchema(typ)
	if !exist {
		logrus.WithFields(logrus.Fields{"type": typ}).Warning("Unable to add human readable attributes, no schema found")
		return
	}
	(*metadata).HumanReadableAttributesFunc = humanReadableAttributesFunc
}

func (r *SchemaRepository) SetDiscriminantFunc(typ string, fn func(self, res *resource.Resource) bool) {
	metadata, exist := r.GetSchema(typ)
	if !exist {
		logrus.WithFields(logrus.Fields{"type": typ}).Warning("Unable to set discriminant function, no schema found")
		return
	}
	(*metadata).DiscriminantFunc = fn
}

package resource

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

type ResourceFactory interface {
	CreateAbstractResource(ty, id string, data map[string]interface{}) *resource.Resource
}

type DriftctlResourceFactory struct {
	resourceSchemaRepository SchemaRepositoryInterface
}

func NewDriftctlResourceFactory(resourceSchemaRepository SchemaRepositoryInterface) *DriftctlResourceFactory {
	return &DriftctlResourceFactory{
		resourceSchemaRepository: resourceSchemaRepository,
	}
}

func (r *DriftctlResourceFactory) CreateAbstractResource(ty, id string, data map[string]interface{}) *resource.Resource {
	attributes := resource.Attributes(data)
	attributes.SanitizeDefaults()

	schema, _ := r.resourceSchemaRepository.GetSchema(ty)
	res := resource.Resource{
		Id:    id,
		Type:  ty,
		Attrs: &attributes,
		Sch:   schema,
	}

	schema, exist := r.resourceSchemaRepository.GetSchema(ty)
	if exist && schema.NormalizeFunc != nil {
		schema.NormalizeFunc(&res)
	}

	return &res
}

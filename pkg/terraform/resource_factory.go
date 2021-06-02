package terraform

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

type TerraformResourceFactory struct {
	resourceSchemaRepository resource.SchemaRepositoryInterface
}

func NewTerraformResourceFactory(resourceSchemaRepository resource.SchemaRepositoryInterface) *TerraformResourceFactory {
	return &TerraformResourceFactory{
		resourceSchemaRepository: resourceSchemaRepository,
	}
}

func (r *TerraformResourceFactory) CreateAbstractResource(ty, id string, data map[string]interface{}) *resource.AbstractResource {
	attributes := resource.Attributes(data)
	attributes.SanitizeDefaults()

	schema, _ := r.resourceSchemaRepository.GetSchema(ty)
	res := resource.AbstractResource{
		Id:    id,
		Type:  ty,
		Attrs: &attributes,
		Sch:   schema,
	}

	schema, exist := r.resourceSchemaRepository.(*resource.SchemaRepository).GetSchema(ty)
	if exist && schema.NormalizeFunc != nil {
		schema.NormalizeFunc(&res)
	}

	return &res
}

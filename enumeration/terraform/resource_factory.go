package terraform

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

type TerraformResourceFactory struct{}

func NewTerraformResourceFactory() *TerraformResourceFactory {
	return &TerraformResourceFactory{}
}

func (r *TerraformResourceFactory) CreateAbstractResource(ty, id string, data map[string]interface{}) *resource.Resource {
	attributes := resource.Attributes(data)

	res := resource.Resource{
		Id:    id,
		Type:  ty,
		Attrs: &attributes,
	}

	return &res
}

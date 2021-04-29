package terraform

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type TerraformResourceFactory struct {
	providerLibrary          *ProviderLibrary
	resourceSchemaRepository *resource.SchemaRepository
}

func NewTerraformResourceFactory(providerLibrary *ProviderLibrary, resourceSchemaRepository *resource.SchemaRepository) *TerraformResourceFactory {
	return &TerraformResourceFactory{
		providerLibrary:          providerLibrary,
		resourceSchemaRepository: resourceSchemaRepository,
	}
}

func (r *TerraformResourceFactory) resolveType(ty string) (cty.Type, error) {
	provider, err := r.providerLibrary.GetProviderForResourceType(ty)
	if err != nil {
		return cty.NilType, err
	}
	if schemas, exist := provider.Schema()[ty]; exist {
		return schemas.Block.ImpliedType(), nil
	}

	return cty.NilType, errors.New("Unable to find ")
}

func (r *TerraformResourceFactory) CreateResource(data interface{}, ty string) (*cty.Value, error) {
	ctyType, err := r.resolveType(ty)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"type": ty,
	}).Debug("Found cty type for resource")

	val, err := gocty.ToCtyValue(data, ctyType)
	if err != nil {
		return nil, err
	}

	return &val, nil
}

func (r *TerraformResourceFactory) CreateAbstractResource(data map[string]interface{}, id, ty string) resource.AbstractResource {
	ctyAttr := resource.Attributes(data)
	ctyAttr.SanitizeDefaultsV3()

	schema, exist := r.resourceSchemaRepository.GetSchema(ty)
	if exist && schema.NormalizeFunc != nil {
		schema.NormalizeFunc(&ctyAttr)
	}

	return resource.AbstractResource{
		Id:    id,
		Type:  ty,
		Attrs: &ctyAttr,
	}
}

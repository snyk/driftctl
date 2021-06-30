package common

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type DetailsFetcher interface {
	ReadDetails(resource.Resource) (resource.Resource, error)
}

type GenericDetailFetcher struct {
	resType      resource.ResourceType
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewGenericDetailFetcher(resType resource.ResourceType, provider terraform.ResourceReader, deserializer *resource.Deserializer) *GenericDetailFetcher {
	return &GenericDetailFetcher{
		resType:      resType,
		reader:       provider,
		deserializer: deserializer,
	}
}

func (f *GenericDetailFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	ctyVal, err := f.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: f.resType,
		ID: res.TerraformId(),
	})
	if err != nil {
		return nil, err
	}
	deserializedRes, err := f.deserializer.DeserializeOne(string(f.resType), *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}

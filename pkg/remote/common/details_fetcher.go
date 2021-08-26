package common

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
)

type DetailsFetcher interface {
	ReadDetails(*resource.Resource) (*resource.Resource, error)
}

type GenericDetailsFetcher struct {
	resType      resource.ResourceType
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewGenericDetailsFetcher(resType resource.ResourceType, provider terraform.ResourceReader, deserializer *resource.Deserializer) *GenericDetailsFetcher {
	return &GenericDetailsFetcher{
		resType:      resType,
		reader:       provider,
		deserializer: deserializer,
	}
}

func (f *GenericDetailsFetcher) ReadDetails(res *resource.Resource) (*resource.Resource, error) {
	attributes := map[string]string{}
	if res.Schema().ResolveReadAttributesFunc != nil {
		attributes = res.Schema().ResolveReadAttributesFunc(res)
	}
	ctyVal, err := f.reader.ReadResource(terraform.ReadResourceArgs{
		Ty:         f.resType,
		ID:         res.ResourceId(),
		Attributes: attributes,
	})
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, res.ResourceType(), res.ResourceId())
	}
	if ctyVal.IsNull() {
		logrus.WithFields(logrus.Fields{
			"type": f.resType,
			"id":   res.ResourceId(),
		}).Debug("Got null while reading resource details")
		return nil, nil
	}
	deserializedRes, err := f.deserializer.DeserializeOne(string(f.resType), *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}

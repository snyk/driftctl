package common

import (
	"strconv"

	"github.com/hashicorp/terraform/flatmap"
	"github.com/sirupsen/logrus"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/terraform"
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
	if res.Attributes() != nil {
		for k, v := range *res.Attributes() {
			if b, ok := v.(bool); ok {
				attributes[k] = strconv.FormatBool(b)
			}
			if i, ok := v.(int); ok {
				attributes[k] = strconv.Itoa(i)
			}
			if i64, ok := v.(int64); ok {
				attributes[k] = strconv.FormatInt(i64, 10)
			}
			if str, ok := v.(string); ok {
				attributes[k] = str
			}
			if sliceOfInterface, ok := v.([]interface{}); ok {
				m := flatmap.Flatten(map[string]interface{}{k: sliceOfInterface})
				for k2, v2 := range m {
					attributes[k2] = v2
				}
			}
		}
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

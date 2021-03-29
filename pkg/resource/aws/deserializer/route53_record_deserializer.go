package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type Route53RecordDeserializer struct {
}

func NewRoute53RecordDeserializer() *Route53RecordDeserializer {
	return &Route53RecordDeserializer{}
}

func (s Route53RecordDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsRoute53RecordResourceType
}

func (s Route53RecordDeserializer) Deserialize(recordList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range recordList {
		rawResource := rawResource
		res, err := decodeRoute53Record(&rawResource)
		if err != nil {
			logrus.Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeRoute53Record(raw *cty.Value) (*resourceaws.AwsRoute53Record, error) {
	var decoded resourceaws.AwsRoute53Record
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = raw
	return &decoded, nil
}

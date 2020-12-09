package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type Route53ZoneDeserializer struct {
}

func NewRoute53ZoneDeserializer() *Route53ZoneDeserializer {
	return &Route53ZoneDeserializer{}
}

func (s Route53ZoneDeserializer) HandledType() resource.ResourceType {
	return aws.AwsRoute53ZoneResourceType
}

func (s Route53ZoneDeserializer) Deserialize(zoneList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawZone := range zoneList {
		zone, err := decodeRoute53Zone(&rawZone)
		if err != nil {
			logrus.Warnf("error when reading zone %+v : %+v", rawZone, err)
		}
		resources = append(resources, zone)
	}
	return resources, nil
}

func decodeRoute53Zone(zone *cty.Value) (resource.Resource, error) {
	var decodedZone aws.AwsRoute53Zone
	if err := gocty.FromCtyValue(*zone, &decodedZone); err != nil {
		return nil, err
	}

	return &decodedZone, nil
}

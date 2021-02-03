package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type SqsQueueDeserializer struct {
}

func NewSqsQueueDeserializer() *SqsQueueDeserializer {
	return &SqsQueueDeserializer{}
}

func (s *SqsQueueDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsSqsQueueResourceType
}

func (s SqsQueueDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		resource, err := decodeSqsQueue(&rawResource)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"type": s.HandledType(),
			}).Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func decodeSqsQueue(raw *cty.Value) (*resourceaws.AwsSqsQueue, error) {
	var decoded resourceaws.AwsSqsQueue
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}

package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type SqsQueuePolicyDeserializer struct {
}

func NewSqsQueuePolicyDeserializer() *SqsQueuePolicyDeserializer {
	return &SqsQueuePolicyDeserializer{}
}

func (s *SqsQueuePolicyDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsSqsQueuePolicyResourceType
}

func (s SqsQueuePolicyDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		resource, err := decodeSqsQueuePolicy(&rawResource)
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

func decodeSqsQueuePolicy(raw *cty.Value) (*resourceaws.AwsSqsQueuePolicy, error) {
	var decoded resourceaws.AwsSqsQueuePolicy
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}

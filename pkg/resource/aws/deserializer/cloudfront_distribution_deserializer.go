package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type CloudfrontDistributionDeserializer struct {
}

func NewCloudfrontDistributionDeserializer() *CloudfrontDistributionDeserializer {
	return &CloudfrontDistributionDeserializer{}
}

func (s *CloudfrontDistributionDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsCloudfrontDistributionResourceType
}

func (s CloudfrontDistributionDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeCloudfrontDistribution(&rawResource)
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

func decodeCloudfrontDistribution(raw *cty.Value) (*resourceaws.AwsCloudfrontDistribution, error) {
	var decoded resourceaws.AwsCloudfrontDistribution
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}

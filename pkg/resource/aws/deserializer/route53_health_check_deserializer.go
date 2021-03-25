package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type Route53HealthCheckDeserializer struct {
}

func NewRoute53HealthCheckDeserializer() *Route53HealthCheckDeserializer {
	return &Route53HealthCheckDeserializer{}
}

func (s *Route53HealthCheckDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsRoute53HealthCheckResourceType
}

func (s Route53HealthCheckDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeRoute53HealthCheck(&rawResource)
		if err != nil {
			logrus.Warnf("Error when deserializing resource %+v : %+v", rawResource, err)
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func decodeRoute53HealthCheck(raw *cty.Value) (*resourceaws.AwsRoute53HealthCheck, error) {
	var decoded resourceaws.AwsRoute53HealthCheck
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	decoded.CtyVal = raw
	return &decoded, nil
}

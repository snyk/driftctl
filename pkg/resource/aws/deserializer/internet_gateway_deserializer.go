package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type InternetGatewayDeserializer struct {
}

func NewInternetGatewayDeserializer() *InternetGatewayDeserializer {
	return &InternetGatewayDeserializer{}
}

func (s *InternetGatewayDeserializer) HandledType() resource.ResourceType {
	return resourceaws.AwsInternetGatewayResourceType
}

func (s InternetGatewayDeserializer) Deserialize(rawList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource
		resource, err := decodeInternetGateway(&rawResource)
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

func decodeInternetGateway(raw *cty.Value) (*resourceaws.AwsInternetGateway, error) {
	var decoded resourceaws.AwsInternetGateway
	if err := gocty.FromCtyValue(*raw, &decoded); err != nil {
		return nil, err
	}
	return &decoded, nil
}

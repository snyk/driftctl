package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type LambdaFunctionDeserializer struct {
}

func NewLambdaFunctionDeserializer() *LambdaFunctionDeserializer {
	return &LambdaFunctionDeserializer{}
}

func (s LambdaFunctionDeserializer) HandledType() resource.ResourceType {
	return aws.AwsLambdaFunctionResourceType
}

func (s LambdaFunctionDeserializer) Deserialize(functionList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawFunction := range functionList {
		function, err := decodeLambdaFunction(rawFunction)
		if err != nil {
			logrus.Warnf("error when reading function %s : %+v", function, err)
			return nil, err
		}
		resources = append(resources, function)
	}
	return resources, nil
}

func decodeLambdaFunction(rawFunction cty.Value) (resource.Resource, error) {
	var decodedFunction aws.AwsLambdaFunction
	if err := gocty.FromCtyValue(rawFunction, &decodedFunction); err != nil {
		return nil, err
	}
	return &decodedFunction, nil
}

package aws

import (
	"github.com/aws/aws-sdk-go/service/apigateway/apigatewayiface"
)

type FakeApiGateway interface {
	apigatewayiface.APIGatewayAPI
}

package aws

import (
	"github.com/aws/aws-sdk-go/service/apigatewayv2/apigatewayv2iface"
)

type FakeApiGatewayV2 interface {
	apigatewayv2iface.ApiGatewayV2API
}

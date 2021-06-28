package aws

import "github.com/aws/aws-sdk-go/service/lambda/lambdaiface"

type FakeLambda interface {
	lambdaiface.LambdaAPI
}

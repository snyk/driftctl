package aws

import "github.com/aws/aws-sdk-go/service/kms/kmsiface"

type FakeKMS interface {
	kmsiface.KMSAPI
}

package aws

import "github.com/aws/aws-sdk-go/service/ecr/ecriface"

type FakeECR interface {
	ecriface.ECRAPI
}

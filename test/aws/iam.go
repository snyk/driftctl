package aws

import (
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

type FakeIAM interface {
	iamiface.IAMAPI
}

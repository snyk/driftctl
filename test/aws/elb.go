package aws

import (
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
)

type FakeELB interface {
	elbiface.ELBAPI
}

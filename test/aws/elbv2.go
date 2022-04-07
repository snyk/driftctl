package aws

import (
	"github.com/aws/aws-sdk-go/service/elbv2/elbv2iface"
)

type FakeELBV2 interface {
	elbv2iface.ELBV2API
}

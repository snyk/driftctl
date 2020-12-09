package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type FakeEC2 interface {
	ec2iface.EC2API
}

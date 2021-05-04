package aws

import "github.com/aws/aws-sdk-go/service/route53/route53iface"

type FakeRoute53 interface {
	route53iface.Route53API
}

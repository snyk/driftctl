package aws

import (
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
)

type FakeSNS interface {
	snsiface.SNSAPI
}

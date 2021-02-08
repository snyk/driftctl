package aws

import (
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

type FakeSQS interface {
	sqsiface.SQSAPI
}

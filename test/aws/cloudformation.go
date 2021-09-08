package aws

import (
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

type FakeCloudformation interface {
	cloudformationiface.CloudFormationAPI
}

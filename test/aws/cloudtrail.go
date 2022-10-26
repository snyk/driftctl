package aws

import (
	"github.com/aws/aws-sdk-go/service/cloudtrail/cloudtrailiface"
)

type FakeCloudtrail interface {
	cloudtrailiface.CloudTrailAPI
}

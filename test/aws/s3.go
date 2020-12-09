package aws

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type FakeS3 interface {
	s3iface.S3API
}

type FakeRequestFailure interface {
	s3.RequestFailure
}

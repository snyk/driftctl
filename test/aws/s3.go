package aws

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3control/s3controliface"
)

type FakeS3 interface {
	s3iface.S3API
}

type FakeS3Control interface {
	s3controliface.S3ControlAPI
}

type FakeRequestFailure interface {
	s3.RequestFailure
}

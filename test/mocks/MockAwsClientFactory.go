package mocks

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type MockAwsClientFactory struct {
	client s3iface.S3API
}

func NewMockAwsClientFactory(client s3iface.S3API) MockAwsClientFactory {
	return MockAwsClientFactory{client}
}

func (s MockAwsClientFactory) GetS3Client(configs ...*aws.Config) s3iface.S3API {
	return s.client
}

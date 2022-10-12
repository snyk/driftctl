package client

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/aws/aws-sdk-go/service/s3control/s3controliface"
)

type AwsClientFactoryInterface interface {
	GetS3Client(configs ...*aws.Config) s3iface.S3API
	GetS3ControlClient(configs ...*aws.Config) s3controliface.S3ControlAPI
}

type AwsClientFactory struct {
	config client.ConfigProvider
}

func NewAWSClientFactory(config client.ConfigProvider) *AwsClientFactory {
	return &AwsClientFactory{config}
}

func (s AwsClientFactory) GetS3Client(configs ...*aws.Config) s3iface.S3API {
	return s3.New(s.config, configs...)
}

func (s AwsClientFactory) GetS3ControlClient(configs ...*aws.Config) s3controliface.S3ControlAPI {
	return s3control.New(s.config, configs...)
}

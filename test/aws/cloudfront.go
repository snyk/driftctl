package aws

import "github.com/aws/aws-sdk-go/service/cloudfront/cloudfrontiface"

type FakeCloudFront interface {
	cloudfrontiface.CloudFrontAPI
}

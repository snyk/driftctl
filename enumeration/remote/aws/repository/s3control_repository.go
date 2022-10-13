package repository

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/snyk/driftctl/enumeration/remote/aws/client"
	"github.com/snyk/driftctl/enumeration/remote/cache"
)

type S3ControlRepository interface {
	DescribeAccountPublicAccessBlock(accountID string) (*s3control.PublicAccessBlockConfiguration, error)
}

type s3ControlRepository struct {
	clientFactory client.AwsClientFactoryInterface
	cache         cache.Cache
}

func NewS3ControlRepository(factory client.AwsClientFactoryInterface, c cache.Cache) *s3ControlRepository {
	return &s3ControlRepository{
		clientFactory: factory,
		cache:         c,
	}
}

func (s *s3ControlRepository) DescribeAccountPublicAccessBlock(accountID string) (*s3control.PublicAccessBlockConfiguration, error) {
	cacheKey := "S3DescribeAccountPublicAccessBlock"
	if v := s.cache.Get(cacheKey); v != nil {
		return v.(*s3control.PublicAccessBlockConfiguration), nil
	}
	out, err := s.clientFactory.GetS3ControlClient(nil).GetPublicAccessBlock(&s3control.GetPublicAccessBlockInput{
		AccountId: aws.String(accountID),
	})

	if err != nil {
		if s.shouldSuppressError(err) {
			return nil, nil
		}

		return nil, err
	}

	result := out.PublicAccessBlockConfiguration

	s.cache.Put(cacheKey, result)
	return result, nil
}

func (s *s3ControlRepository) shouldSuppressError(err error) bool {
	if requestFailure, ok := err.(awserr.RequestFailure); ok {
		if requestFailure.Code() == "NoSuchPublicAccessBlockConfiguration" {
			// do not throw the error up if there is no access block config
			return true
		}
	}
	return false
}

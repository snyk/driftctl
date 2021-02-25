package enumerator

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/cloudskiff/driftctl/pkg/iac/config"
	"github.com/pkg/errors"
)

type S3EnumeratorConfig struct {
	Bucket *string
	Prefix *string
}

type S3Enumerator struct {
	config config.SupplierConfig
	client s3iface.S3API
}

func NewS3Enumerator(config config.SupplierConfig) *S3Enumerator {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return &S3Enumerator{
		config,
		s3.New(sess),
	}
}

func (s *S3Enumerator) Enumerate() ([]string, error) {
	bucketPath := strings.Split(s.config.Path, "/")
	if len(bucketPath) < 2 {
		return nil, errors.Errorf("Unable to parse S3 path: %s. Must be BUCKET_NAME/PREFIX", s.config.Path)
	}
	bucket := bucketPath[0]
	prefix := strings.Join(bucketPath[1:], "/")

	keys := make([]string, 0)
	input := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	}
	err := s.client.ListObjectsV2Pages(input, func(output *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, metadata := range output.Contents {
			keys = append(keys, strings.Join([]string{bucket, *metadata.Key}, "/"))
		}
		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return keys, nil
}

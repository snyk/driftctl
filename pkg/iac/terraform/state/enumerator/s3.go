package enumerator

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/cloudskiff/driftctl/pkg/envproxy"
	"github.com/pkg/errors"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/cloudskiff/driftctl/pkg/iac/config"
)

type S3Enumerator struct {
	config config.SupplierConfig
	client s3iface.S3API
}

func NewS3Enumerator(config config.SupplierConfig) *S3Enumerator {
	envProxy := envproxy.NewEnvProxy("DCTL_S3_", "AWS_")
	envProxy.Apply()
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	envProxy.Restore()
	return &S3Enumerator{
		config,
		s3.New(sess),
	}
}

func (s *S3Enumerator) Path() string {
	return s.config.Path
}

func (s *S3Enumerator) Enumerate() ([]string, error) {
	bucketPath := strings.Split(s.config.Path, "/")
	if len(bucketPath) < 2 {
		return nil, errors.Errorf("Unable to parse S3 path: %s. Must be BUCKET_NAME/PREFIX", s.config.Path)
	}

	bucket := bucketPath[0]
	// prefix should contains everything that does not have a glob pattern
	// Pattern should be the glob matcher string
	prefix, pattern := GlobS3(strings.Join(bucketPath[1:], "/"))

	fullPattern := strings.Join([]string{prefix, pattern}, "/")
	fullPattern = strings.Trim(fullPattern, "/")

	files := make([]string, 0)
	input := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &prefix,
	}
	err := s.client.ListObjectsV2Pages(input, func(output *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, metadata := range output.Contents {
			if aws.Int64Value(metadata.Size) > 0 {
				key := *metadata.Key
				if match, _ := doublestar.Match(fullPattern, key); match {
					files = append(files, strings.Join([]string{bucket, key}, "/"))
				}
			}
		}
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return files, fmt.Errorf("no Terraform state was found in %s, exiting", s.config.Path)
	}

	return files, nil
}

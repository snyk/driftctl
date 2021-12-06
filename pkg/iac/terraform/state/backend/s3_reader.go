package backend

import (
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/envproxy"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

const BackendKeyS3 = "s3"

type S3Backend struct {
	input    s3.GetObjectInput
	reader   io.ReadCloser
	S3Client s3iface.S3API
}

func NewS3Reader(path string) (*S3Backend, error) {

	backend := S3Backend{}
	bucketPath := strings.Split(path, "/")
	if len(bucketPath) < 2 {
		return nil, errors.Errorf("Unable to parse S3 path: %s. Must be BUCKET_NAME/PATH/TO/OBJECT", path)
	}
	bucket := bucketPath[0]
	key := strings.Join(bucketPath[1:], "/")

	backend.input = s3.GetObjectInput{
		Key:    &key,
		Bucket: &bucket,
	}
	envProxy := envproxy.NewEnvProxy("DCTL_S3_", "AWS_")
	envProxy.Apply()
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	envProxy.Restore()
	backend.S3Client = s3.New(sess)
	return &backend, nil
}

func (s *S3Backend) Read(p []byte) (n int, err error) {
	if s.reader == nil {
		response, err := s.S3Client.GetObject(&s.input)
		if err != nil {
			requestFailure, ok := err.(s3.RequestFailure)
			if ok {
				return 0, errors.Errorf(
					"Error reading state '%s' from s3 bucket '%s': %s",
					*s.input.Key,
					*s.input.Bucket,
					requestFailure.Message(),
				)
			}
			return 0, err
		}
		s.reader = response.Body
	}
	return s.reader.Read(p)
}

func (s *S3Backend) Close() error {
	if s.reader != nil {
		return s.reader.Close()
	}
	return errors.New("Unable to close reader as nothing was opened")
}

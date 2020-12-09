package aws

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/cloudskiff/driftctl/pkg"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
)

type S3BucketSupplier struct {
	reader           terraform.ResourceReader
	deserializer     deserializer.CTYDeserializer
	awsClientFactory AwsClientFactoryInterface
	runner           *terraform.ParallelResourceReader
}

func NewS3BucketSupplier(runner *pkg.ParallelRunner, factory AwsClientFactoryInterface) *S3BucketSupplier {
	return &S3BucketSupplier{terraform.Provider(terraform.AWS), awsdeserializer.NewS3BucketDeserializer(), factory, terraform.NewParallelResourceReader(runner)}
}

func (s S3BucketSupplier) Resources() ([]resource.Resource, error) {
	retrieve, err := s.list()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(retrieve)
}

func (s *S3BucketSupplier) list() ([]cty.Value, error) {
	input := &s3.ListBucketsInput{}

	s3Client := s.awsClientFactory.GetS3Client(nil)

	response, err := s3Client.ListBuckets(input)
	if err != nil {
		return nil, err
	}

	for _, bucket := range response.Buckets {
		b := *bucket

		s.runner.Run(func() (cty.Value, error) {
			return s.readBucket(b, &s3Client)
		})
	}
	return s.runner.Wait()
}

func readBucketRegion(client *s3iface.S3API, name string) (string, error) {
	bucketLocationRequest := s3.GetBucketLocationInput{Bucket: &name}
	bucketLocationResponse, err := (*client).GetBucketLocation(&bucketLocationRequest)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == s3.ErrCodeNoSuchBucket {
			logrus.WithFields(logrus.Fields{
				"bucket": name,
			}).Warning("Unable to retrieve bucket region, this may be an inconsistency in S3 api for fresh deleted bucket, skipping ...")
			return "", nil
		}
		return "", err
	}
	return *bucketLocationResponse.LocationConstraint, nil
}

func (s *S3BucketSupplier) readBucket(bucket s3.Bucket, client *s3iface.S3API) (cty.Value, error) {
	name := *bucket.Name

	region, err := readBucketRegion(client, name)
	if err != nil {
		return cty.NilVal, err
	}
	if region == "" {
		return cty.NilVal, nil
	}

	s3Bucket, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: aws.AwsS3BucketResourceType,
		ID: name,
		Attributes: map[string]string{
			"aws_region": region,
		},
	})
	if err != nil {
		return cty.NilVal, err
	}
	return *s3Bucket, err
}

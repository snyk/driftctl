package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/zclconf/go-cty/cty"
)

type S3BucketNotificationSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	factory      AwsClientFactoryInterface
	runner       *terraform.ParallelResourceReader
}

func NewS3BucketNotificationSupplier(provider *TerraformProvider, factory AwsClientFactoryInterface) *S3BucketNotificationSupplier {
	return &S3BucketNotificationSupplier{
		provider,
		awsdeserializer.NewS3BucketNotificationDeserializer(),
		factory, terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *S3BucketNotificationSupplier) Resources() ([]resource.Resource, error) {
	input := &s3.ListBucketsInput{}

	client := s.factory.GetS3Client(nil)
	response, err := client.ListBuckets(input)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, aws.AwsS3BucketNotificationResourceType, aws.AwsS3BucketResourceType)
	}

	for _, bucket := range response.Buckets {
		name := *bucket.Name
		region, err := readBucketRegion(&client, name)
		if err != nil {
			return nil, err
		}
		if region == "" {
			continue
		}
		s.listBucketNotificationConfiguration(*bucket.Name, region)
	}
	ctyVals, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}
	deserializedValues, err := s.deserializer.Deserialize(ctyVals)
	results := make([]resource.Resource, 0, len(deserializedValues))
	if err != nil {
		return deserializedValues, err
	}
	for _, val := range deserializedValues {
		res, ok := val.(*aws.AwsS3BucketNotification)
		if ok {
			if (res.LambdaFunction != nil && len(*res.LambdaFunction) > 0) ||
				(res.Queue != nil && len(*res.Queue) > 0) ||
				(res.Topic != nil && len(*res.Topic) > 0) {
				results = append(results, res)
			}
		}
	}
	return results, nil
}

func (s *S3BucketNotificationSupplier) listBucketNotificationConfiguration(name, region string) {
	s.runner.Run(func() (cty.Value, error) {
		s3BucketPolicy, err := s.reader.ReadResource(terraform.ReadResourceArgs{
			Ty: aws.AwsS3BucketNotificationResourceType,
			ID: name,
			Attributes: map[string]string{
				"aws_region": region,
			},
		})
		if err != nil {
			return cty.NilVal, err
		}
		return *s3BucketPolicy, err
	})
}

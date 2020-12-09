package aws

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cloudskiff/driftctl/pkg"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/zclconf/go-cty/cty"
)

type S3BucketPolicySupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	factory      AwsClientFactoryInterface
	runner       *terraform.ParallelResourceReader
}

func NewS3BucketPolicySupplier(runner *pkg.ParallelRunner, factory AwsClientFactoryInterface) *S3BucketPolicySupplier {
	return &S3BucketPolicySupplier{terraform.Provider(terraform.AWS), awsdeserializer.NewS3BucketPolicyDeserializer(), factory, terraform.NewParallelResourceReader(runner)}
}

func (s *S3BucketPolicySupplier) Resources() ([]resource.Resource, error) {
	input := &s3.ListBucketsInput{}

	client := s.factory.GetS3Client(nil)
	response, err := client.ListBuckets(input)
	if err != nil {
		return nil, err
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
		s.listBucketPolicyConfiguration(*bucket.Name, region)
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
		res, ok := val.(*aws.AwsS3BucketPolicy)
		if ok {
			if res.Policy != nil && *res.Policy != "" {
				results = append(results, res)
			}
		}
	}
	return results, nil
}

func (s *S3BucketPolicySupplier) listBucketPolicyConfiguration(name, region string) {
	s.runner.Run(func() (cty.Value, error) {
		s3BucketNotification, err := s.reader.ReadResource(
			terraform.ReadResourceArgs{
				Ty: aws.AwsS3BucketPolicyResourceType,
				ID: name,
				Attributes: map[string]string{
					"aws_region": region,
				},
			},
		)

		if err != nil {
			return cty.NilVal, err
		}
		return *s3BucketNotification, err
	})
}

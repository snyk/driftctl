package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/zclconf/go-cty/cty"
)

type S3BucketPolicySupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	repository   repository.S3Repository
	runner       *terraform.ParallelResourceReader
}

func NewS3BucketPolicySupplier(provider *AWSTerraformProvider, repository repository.S3Repository) *S3BucketPolicySupplier {
	return &S3BucketPolicySupplier{
		provider,
		awsdeserializer.NewS3BucketPolicyDeserializer(),
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *S3BucketPolicySupplier) Resources() ([]resource.Resource, error) {
	buckets, err := s.repository.ListAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, aws.AwsS3BucketPolicyResourceType, aws.AwsS3BucketResourceType)
	}

	for _, bucket := range buckets {
		bucket := *bucket
		region, err := s.repository.GetBucketLocation(&bucket)
		if err != nil {
			return nil, err
		}
		if region == "" {
			continue
		}
		s.runner.Run(func() (cty.Value, error) {
			s3BucketNotification, err := s.reader.ReadResource(
				terraform.ReadResourceArgs{
					Ty: aws.AwsS3BucketPolicyResourceType,
					ID: *bucket.Name,
					Attributes: map[string]string{
						"alias": region,
					},
				},
			)

			if err != nil {
				return cty.NilVal, err
			}
			return *s3BucketNotification, err
		})
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

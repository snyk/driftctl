package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	tf "github.com/cloudskiff/driftctl/pkg/remote/terraform"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/zclconf/go-cty/cty"
)

type S3BucketNotificationSupplier struct {
	reader         terraform.ResourceReader
	deserializer   *resource.Deserializer
	repository     repository.S3Repository
	runner         *terraform.ParallelResourceReader
	providerConfig tf.TerraformProviderConfig
}

func NewS3BucketNotificationSupplier(provider *AWSTerraformProvider, repository repository.S3Repository, deserializer *resource.Deserializer) *S3BucketNotificationSupplier {
	return &S3BucketNotificationSupplier{
		provider,
		deserializer,
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
		provider.Config,
	}
}

func (s *S3BucketNotificationSupplier) Resources() ([]resource.Resource, error) {
	buckets, err := s.repository.ListAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, aws.AwsS3BucketNotificationResourceType, aws.AwsS3BucketResourceType)
	}

	for _, bucket := range buckets {
		bucket := *bucket
		region, err := s.repository.GetBucketLocation(*bucket.Name)
		if err != nil {
			return nil, err
		}
		if region == "" || region != s.providerConfig.DefaultAlias {
			continue
		}
		s.runner.Run(func() (cty.Value, error) {
			s3BucketPolicy, err := s.reader.ReadResource(terraform.ReadResourceArgs{
				Ty: aws.AwsS3BucketNotificationResourceType,
				ID: *bucket.Name,
				Attributes: map[string]string{
					"alias": region,
				},
			})
			if err != nil {
				return cty.NilVal, err
			}
			return *s3BucketPolicy, err
		})
	}
	ctyVals, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}
	deserializedValues, err := s.deserializer.Deserialize(aws.AwsS3BucketNotificationResourceType, ctyVals)
	results := make([]resource.Resource, 0, len(deserializedValues))
	if err != nil {
		return deserializedValues, err
	}
	for _, val := range deserializedValues {
		res, _ := val.(*resource.AbstractResource)

		if ((*res.Attrs)["lambda_function"] != nil && len((*res.Attrs)["lambda_function"].([]interface{})) > 0) ||
			((*res.Attrs)["queue"] != nil && len((*res.Attrs)["queue"].([]interface{})) > 0) ||
			((*res.Attrs)["topic"] != nil && len((*res.Attrs)["topic"].([]interface{})) > 0) {
			results = append(results, res)
		}

	}
	return results, nil
}

package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	tf "github.com/cloudskiff/driftctl/pkg/remote/terraform"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

type S3BucketPolicyEnumerator struct {
	repository     repository.S3Repository
	factory        resource.ResourceFactory
	providerConfig tf.TerraformProviderConfig
}

func NewS3BucketPolicyEnumerator(repo repository.S3Repository, factory resource.ResourceFactory, providerConfig tf.TerraformProviderConfig) *S3BucketPolicyEnumerator {
	return &S3BucketPolicyEnumerator{
		repository:     repo,
		factory:        factory,
		providerConfig: providerConfig,
	}
}

func (e *S3BucketPolicyEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsS3BucketPolicyResourceType
}

func (e *S3BucketPolicyEnumerator) Enumerate() ([]resource.Resource, error) {
	buckets, err := e.repository.ListAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, string(e.SupportedType()), aws.AwsS3BucketResourceType)
	}

	results := make([]resource.Resource, len(buckets))

	for _, bucket := range buckets {
		region, err := e.repository.GetBucketLocation(*bucket.Name)
		if err != nil {
			return nil, err
		}
		if region == "" || region != e.providerConfig.DefaultAlias {
			logrus.WithFields(logrus.Fields{
				"region": region,
				"bucket": *bucket.Name,
			}).Debug("Skipped bucket policy")
			continue
		}

		policy, err := e.repository.GetBucketPolicy(*bucket.Name, region)
		if err != nil {
			return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsS3BucketPolicyResourceType)
		}

		if policy != nil {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*bucket.Name,
					map[string]interface{}{
						"region": region,
					},
				),
			)
		}
	}

	return results, err
}

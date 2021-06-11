package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	tf "github.com/cloudskiff/driftctl/pkg/remote/terraform"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

type S3BucketEnumerator struct {
	repository     repository.S3Repository
	factory        resource.ResourceFactory
	providerConfig tf.TerraformProviderConfig
}

func NewS3BucketEnumerator(repo repository.S3Repository, factory resource.ResourceFactory, providerConfig tf.TerraformProviderConfig) *S3BucketEnumerator {
	return &S3BucketEnumerator{
		repository:     repo,
		factory:        factory,
		providerConfig: providerConfig,
	}
}

func (e *S3BucketEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsS3BucketResourceType
}

func (e *S3BucketEnumerator) Enumerate() ([]resource.Resource, error) {
	buckets, err := e.repository.ListAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, string(e.SupportedType()))
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
			}).Debug("Skipped bucket")
			continue
		}
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

	return results, err
}

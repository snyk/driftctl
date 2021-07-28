package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	tf "github.com/cloudskiff/driftctl/pkg/remote/terraform"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

type S3BucketNotificationEnumerator struct {
	repository     repository.S3Repository
	factory        resource.ResourceFactory
	providerConfig tf.TerraformProviderConfig
}

func NewS3BucketNotificationEnumerator(repo repository.S3Repository, factory resource.ResourceFactory, providerConfig tf.TerraformProviderConfig) *S3BucketNotificationEnumerator {
	return &S3BucketNotificationEnumerator{
		repository:     repo,
		factory:        factory,
		providerConfig: providerConfig,
	}
}

func (e *S3BucketNotificationEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsS3BucketNotificationResourceType
}

func (e *S3BucketNotificationEnumerator) Enumerate() ([]resource.Resource, error) {
	buckets, err := e.repository.ListAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceScanningErrorWithType(err, string(e.SupportedType()), aws.AwsS3BucketResourceType)
	}

	results := make([]resource.Resource, len(buckets))

	for _, bucket := range buckets {
		region, err := e.repository.GetBucketLocation(*bucket.Name)
		if err != nil {
			return nil, remoteerror.NewResourceScanningErrorWithType(err, string(e.SupportedType()), aws.AwsS3BucketResourceType)
		}
		if region == "" || region != e.providerConfig.DefaultAlias {
			logrus.WithFields(logrus.Fields{
				"region": region,
				"bucket": *bucket.Name,
			}).Debug("Skipped bucket")
			continue
		}

		notification, err := e.repository.GetBucketNotification(*bucket.Name, region)
		if err != nil {
			return nil, remoteerror.NewResourceScanningError(err, string(e.SupportedType()))
		}

		if notification == nil {
			logrus.WithFields(logrus.Fields{
				"region": region,
				"bucket": *bucket.Name,
			}).Debug("Skipped empty bucket notification")
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

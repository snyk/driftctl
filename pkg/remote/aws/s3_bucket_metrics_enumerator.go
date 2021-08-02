package aws

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	tf "github.com/cloudskiff/driftctl/pkg/remote/terraform"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

type S3BucketMetricsEnumerator struct {
	repository     repository.S3Repository
	factory        resource.ResourceFactory
	providerConfig tf.TerraformProviderConfig
}

func NewS3BucketMetricsEnumerator(repo repository.S3Repository, factory resource.ResourceFactory, providerConfig tf.TerraformProviderConfig) *S3BucketMetricsEnumerator {
	return &S3BucketMetricsEnumerator{
		repository:     repo,
		factory:        factory,
		providerConfig: providerConfig,
	}
}

func (e *S3BucketMetricsEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsS3BucketMetricResourceType
}

func (e *S3BucketMetricsEnumerator) Enumerate() ([]resource.Resource, error) {
	buckets, err := e.repository.ListAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceScanningErrorWithType(err, aws.AwsS3BucketMetricResourceType, aws.AwsS3BucketResourceType)
	}

	results := make([]resource.Resource, len(buckets))

	for _, bucket := range buckets {
		region, err := e.repository.GetBucketLocation(*bucket.Name)
		if err != nil {
			return nil, remoteerror.NewResourceScanningErrorWithType(err, aws.AwsS3BucketMetricResourceType, aws.AwsS3BucketResourceType)
		}
		if region == "" || region != e.providerConfig.DefaultAlias {
			logrus.WithFields(logrus.Fields{
				"region": region,
				"bucket": *bucket.Name,
			}).Debug("Skipped bucket")
			continue
		}

		metricsConfigurationList, err := e.repository.ListBucketMetricsConfigurations(bucket, region)
		if err != nil {
			return nil, remoteerror.NewResourceScanningError(err, aws.AwsS3BucketMetricResourceType)
		}

		for _, metric := range metricsConfigurationList {
			id := fmt.Sprintf("%s:%s", *bucket.Name, *metric.Id)
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					id,
					map[string]interface{}{
						"region": region,
					},
				),
			)
		}
	}

	return results, nil
}

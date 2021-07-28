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

type S3BucketInventoryEnumerator struct {
	repository     repository.S3Repository
	factory        resource.ResourceFactory
	providerConfig tf.TerraformProviderConfig
}

func NewS3BucketInventoryEnumerator(repo repository.S3Repository, factory resource.ResourceFactory, providerConfig tf.TerraformProviderConfig) *S3BucketInventoryEnumerator {
	return &S3BucketInventoryEnumerator{
		repository:     repo,
		factory:        factory,
		providerConfig: providerConfig,
	}
}

func (e *S3BucketInventoryEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsS3BucketInventoryResourceType
}

func (e *S3BucketInventoryEnumerator) Enumerate() ([]resource.Resource, error) {
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
			}).Debug("Skipped bucket inventory")
			continue
		}

		inventoryConfigurations, err := e.repository.ListBucketInventoryConfigurations(bucket, region)
		if err != nil {
			return nil, remoteerror.NewResourceScanningError(err, aws.AwsS3BucketInventoryResourceType)
		}

		for _, config := range inventoryConfigurations {
			id := fmt.Sprintf("%s:%s", *bucket.Name, *config.Id)
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

	return results, err
}

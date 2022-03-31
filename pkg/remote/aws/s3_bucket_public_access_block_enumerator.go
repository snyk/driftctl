package aws

import (
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/alerter"
	"github.com/snyk/driftctl/pkg/remote/alerts"
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	"github.com/snyk/driftctl/pkg/remote/common"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	tf "github.com/snyk/driftctl/pkg/remote/terraform"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type S3BucketPublicAccessBlockEnumerator struct {
	repository     repository.S3Repository
	factory        resource.ResourceFactory
	providerConfig tf.TerraformProviderConfig
	alerter        alerter.AlerterInterface
}

func NewS3BucketPublicAccessBlockEnumerator(repo repository.S3Repository, factory resource.ResourceFactory, providerConfig tf.TerraformProviderConfig, alerter alerter.AlerterInterface) *S3BucketPublicAccessBlockEnumerator {
	return &S3BucketPublicAccessBlockEnumerator{
		repository:     repo,
		factory:        factory,
		providerConfig: providerConfig,
		alerter:        alerter,
	}
}

func (e *S3BucketPublicAccessBlockEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsS3BucketPublicAccessBlockResourceType
}

func (e *S3BucketPublicAccessBlockEnumerator) Enumerate() ([]*resource.Resource, error) {
	buckets, err := e.repository.ListAllBuckets()
	if err != nil {
		return nil, remoteerror.NewResourceListingErrorWithType(err, string(e.SupportedType()), aws.AwsS3BucketResourceType)
	}

	results := make([]*resource.Resource, 0, len(buckets))

	for _, bucket := range buckets {
		region, err := e.repository.GetBucketLocation(*bucket.Name)
		if err != nil {
			alerts.SendEnumerationAlert(common.RemoteAWSTerraform, e.alerter, remoteerror.NewResourceScanningError(err, string(e.SupportedType()), *bucket.Name))
			continue
		}
		if region == "" || region != e.providerConfig.DefaultAlias {
			logrus.WithFields(logrus.Fields{
				"region": region,
				"bucket": *bucket.Name,
			}).Debug("Skipped bucket public access block")
			continue
		}

		block, err := e.repository.GetBucketPublicAccessBlock(*bucket.Name, region)
		if err != nil {
			alerts.SendEnumerationAlert(common.RemoteAWSTerraform, e.alerter, remoteerror.NewResourceScanningError(err, string(e.SupportedType()), *bucket.Name))
			continue
		}

		if block != nil {
			results = append(
				results,
				e.factory.CreateAbstractResource(
					string(e.SupportedType()),
					*bucket.Name,
					map[string]interface{}{
						"block_public_acls":       awssdk.BoolValue(block.BlockPublicAcls),
						"block_public_policy":     awssdk.BoolValue(block.BlockPublicPolicy),
						"ignore_public_acls":      awssdk.BoolValue(block.IgnorePublicAcls),
						"restrict_public_buckets": awssdk.BoolValue(block.RestrictPublicBuckets),
					},
				),
			)
		}
	}

	return results, err
}

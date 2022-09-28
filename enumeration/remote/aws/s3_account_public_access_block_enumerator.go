package aws

import (
	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	tf "github.com/snyk/driftctl/enumeration/remote/terraform"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type S3AccountPublicAccessBlockEnumerator struct {
	repository     repository.S3ControlRepository
	factory        resource.ResourceFactory
	providerConfig tf.TerraformProviderConfig
	alerter        alerter.AlerterInterface
}

func NewS3AccountPublicAccessBlockEnumerator(repo repository.S3ControlRepository, factory resource.ResourceFactory, providerConfig tf.TerraformProviderConfig, alerter alerter.AlerterInterface) *S3AccountPublicAccessBlockEnumerator {
	return &S3AccountPublicAccessBlockEnumerator{
		repository:     repo,
		factory:        factory,
		providerConfig: providerConfig,
		alerter:        alerter,
	}
}

func (e *S3AccountPublicAccessBlockEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsS3AccountPublicAccessBlock
}

func (e *S3AccountPublicAccessBlockEnumerator) Enumerate() ([]*resource.Resource, error) {
	accountPublicAccessBlock, err := e.repository.DescribeAccountPublicAccessBlock()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, 1)

	results = append(
		results,
		e.factory.CreateAbstractResource(
			string(e.SupportedType()),
			e.repository.GetAccountID(),
			map[string]interface{}{
				"block_public_acls":       awssdk.BoolValue(accountPublicAccessBlock.BlockPublicAcls),
				"block_public_policy":     awssdk.BoolValue(accountPublicAccessBlock.BlockPublicPolicy),
				"ignore_public_acls":      awssdk.BoolValue(accountPublicAccessBlock.IgnorePublicAcls),
				"restrict_public_buckets": awssdk.BoolValue(accountPublicAccessBlock.RestrictPublicBuckets),
			},
		),
	)

	return results, err
}

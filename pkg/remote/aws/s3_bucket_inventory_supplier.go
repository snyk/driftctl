package aws

import (
	"fmt"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	awssdk "github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type S3BucketInventorySupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	factory      AwsClientFactoryInterface
	runner       *terraform.ParallelResourceReader
}

func NewS3BucketInventorySupplier(provider *TerraformProvider, factory AwsClientFactoryInterface) *S3BucketInventorySupplier {
	return &S3BucketInventorySupplier{
		provider,
		awsdeserializer.NewS3BucketInventoryDeserializer(),
		factory,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *S3BucketInventorySupplier) Resources() ([]resource.Resource, error) {
	input := &s3.ListBucketsInput{}

	client := s.factory.GetS3Client(nil)
	response, err := client.ListBuckets(input)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, aws.AwsS3BucketInventoryResourceType, aws.AwsS3BucketResourceType)
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
		if err := s.listBucketInventoryConfiguration(*bucket.Name, region); err != nil {
			return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsS3BucketInventoryResourceType)
		}
	}
	ctyVals, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(ctyVals)
}

func (s *S3BucketInventorySupplier) listBucketInventoryConfiguration(name, region string) error {
	request := &s3.ListBucketInventoryConfigurationsInput{
		Bucket:            &name,
		ContinuationToken: nil,
	}

	inventoryConfigurations := make([]*s3.InventoryConfiguration, 0)
	client := s.factory.GetS3Client(&awssdk.Config{Region: &region})

	for {
		configurations, err := client.ListBucketInventoryConfigurations(request)
		if err != nil {
			logrus.Warnf("Error listing bucket inventory configuration %s: %+v", name, err)
			return err
		}
		inventoryConfigurations = append(inventoryConfigurations, configurations.InventoryConfigurationList...)
		if configurations.IsTruncated != nil && *configurations.IsTruncated {
			request.ContinuationToken = configurations.NextContinuationToken
		} else {
			break
		}
	}

	for _, config := range inventoryConfigurations {
		id := fmt.Sprintf("%s:%s", name, *config.Id)
		s.runner.Run(func() (cty.Value, error) {
			s3BucketInventory, err := s.reader.ReadResource(
				terraform.ReadResourceArgs{
					Ty: aws.AwsS3BucketInventoryResourceType,
					ID: id,
					Attributes: map[string]string{
						"aws_region": region,
					},
				},
			)

			if err != nil {
				return cty.NilVal, err
			}
			return *s3BucketInventory, err
		})
	}

	return nil
}

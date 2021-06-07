package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type InternetGatewaySupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.EC2Repository
	runner       *terraform.ParallelResourceReader
}

func NewInternetGatewaySupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.EC2Repository) *InternetGatewaySupplier {
	return &InternetGatewaySupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *InternetGatewaySupplier) Resources() ([]resource.Resource, error) {
	internetGateways, err := s.client.ListAllInternetGateways()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsInternetGatewayResourceType)
	}

	for _, internetGateway := range internetGateways {
		gtw := *internetGateway
		s.runner.Run(func() (cty.Value, error) {
			return s.readInternetGateway(gtw)
		})
	}

	resources, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(aws.AwsInternetGatewayResourceType, resources)
}

func (s *InternetGatewaySupplier) readInternetGateway(internetGateway ec2.InternetGateway) (cty.Value, error) {
	var Ty resource.ResourceType = aws.AwsInternetGatewayResourceType
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: Ty,
		ID: *internetGateway.InternetGatewayId,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type": Ty,
		}).Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}

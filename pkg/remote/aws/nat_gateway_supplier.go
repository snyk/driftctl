package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type NatGatewaySupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       repository.EC2Repository
	runner       *terraform.ParallelResourceReader
}

func NewNatGatewaySupplier(provider *AWSTerraformProvider) *NatGatewaySupplier {
	return &NatGatewaySupplier{
		provider,
		awsdeserializer.NewNatGatewayDeserializer(),
		repository.NewEC2Repository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *NatGatewaySupplier) Resources() ([]resource.Resource, error) {
	retrievedNatGateways, err := s.client.ListAllNatGateways()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsNatGatewayResourceType)
	}

	for _, gateway := range retrievedNatGateways {
		res := *gateway
		s.runner.Run(func() (cty.Value, error) {
			return s.readNatGateway(res)
		})
	}

	// Retrieve results from terraform provider
	natGatewayResources, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	// Deserialize
	resources, err := s.deserializer.Deserialize(natGatewayResources)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (s *NatGatewaySupplier) readNatGateway(gateway ec2.NatGateway) (cty.Value, error) {
	var Ty resource.ResourceType = aws.AwsNatGatewayResourceType
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *gateway.NatGatewayId,
		Ty: Ty,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type": Ty,
		}).Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}

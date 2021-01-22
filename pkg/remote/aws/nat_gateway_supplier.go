package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
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
	client       ec2iface.EC2API
	runner       *terraform.ParallelResourceReader
}

func NewNatGatewaySupplier(provider *TerraformProvider) *NatGatewaySupplier {
	return &NatGatewaySupplier{
		provider,
		awsdeserializer.NewNatGatewayDeserializer(),
		ec2.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s NatGatewaySupplier) Resources() ([]resource.Resource, error) {

	retrievedNatGateways, err := listNatGateways(s.client)
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

func (s NatGatewaySupplier) readNatGateway(gateway ec2.NatGateway) (cty.Value, error) {
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

func listNatGateways(client ec2iface.EC2API) ([]*ec2.NatGateway, error) {
	var result []*ec2.NatGateway
	input := ec2.DescribeNatGatewaysInput{}
	err := client.DescribeNatGatewaysPages(&input,
		func(resp *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
			result = append(result, resp.NatGateways...)
			return !lastPage
		},
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

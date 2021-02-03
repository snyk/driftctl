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

type InternetGatewaySupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       ec2iface.EC2API
	runner       *terraform.ParallelResourceReader
}

func NewInternetGatewaySupplier(provider *TerraformProvider) *InternetGatewaySupplier {
	return &InternetGatewaySupplier{
		provider,
		awsdeserializer.NewInternetGatewayDeserializer(),
		ec2.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s InternetGatewaySupplier) Resources() ([]resource.Resource, error) {
	internetGateways, err := listInternetGateways(s.client)
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

	return s.deserializer.Deserialize(resources)
}

func (s InternetGatewaySupplier) readInternetGateway(internetGateway ec2.InternetGateway) (cty.Value, error) {
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

func listInternetGateways(client ec2iface.EC2API) ([]*ec2.InternetGateway, error) {
	var internetGateways []*ec2.InternetGateway
	input := ec2.DescribeInternetGatewaysInput{}
	err := client.DescribeInternetGatewaysPages(&input,
		func(resp *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
			internetGateways = append(internetGateways, resp.InternetGateways...)
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}
	return internetGateways, nil
}

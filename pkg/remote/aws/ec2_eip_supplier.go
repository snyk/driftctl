package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type EC2EipSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       ec2iface.EC2API
	runner       *terraform.ParallelResourceReader
}

func NewEC2EipSupplier(provider *TerraformProvider) *EC2EipSupplier {
	return &EC2EipSupplier{
		provider,
		awsdeserializer.NewEC2EipDeserializer(),
		ec2.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s EC2EipSupplier) Resources() ([]resource.Resource, error) {
	addresses, err := listAddresses(s.client)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsEipResourceType)
	}
	results := make([]cty.Value, 0)
	if len(addresses) > 0 {
		for _, address := range addresses {
			addr := *address
			s.runner.Run(func() (cty.Value, error) {
				return s.readEIP(addr)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(results)
}

func (s EC2EipSupplier) readEIP(address ec2.Address) (cty.Value, error) {
	id := aws.StringValue(address.AllocationId)
	resAddress, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: resourceaws.AwsEipResourceType,
		ID: id,
	})
	if err != nil {
		logrus.Warnf("Error reading eip %s[%s]: %+v", id, resourceaws.AwsEipResourceType, err)
		return cty.NilVal, err
	}
	return *resAddress, nil
}

func listAddresses(client ec2iface.EC2API) ([]*ec2.Address, error) {
	input := &ec2.DescribeAddressesInput{}
	response, err := client.DescribeAddresses(input)
	if err != nil {
		return nil, err
	}
	return response.Addresses, nil
}

package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/cloudskiff/driftctl/pkg"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/zclconf/go-cty/cty"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
)

type SubnetSupplier struct {
	reader                    terraform.ResourceReader
	defaultSubnetDeserializer deserializer.CTYDeserializer
	subnetDeserializer        deserializer.CTYDeserializer
	client                    ec2iface.EC2API
	defaultSubnetRunner       *terraform.ParallelResourceReader
	subnetRunner              *terraform.ParallelResourceReader
}

func NewSubnetSupplier(runner *pkg.ParallelRunner, client ec2iface.EC2API) *SubnetSupplier {
	return &SubnetSupplier{
		terraform.Provider(terraform.AWS),
		awsdeserializer.NewDefaultSubnetDeserializer(),
		awsdeserializer.NewSubnetDeserializer(),
		client,
		terraform.NewParallelResourceReader(runner.SubRunner()),
		terraform.NewParallelResourceReader(runner.SubRunner()),
	}
}

func (s SubnetSupplier) Resources() ([]resource.Resource, error) {
	input := ec2.DescribeSubnetsInput{}
	var subnets []*ec2.Subnet
	var defaultSubnets []*ec2.Subnet
	err := s.client.DescribeSubnetsPages(&input,
		func(resp *ec2.DescribeSubnetsOutput, lastPage bool) bool {
			for _, subnet := range resp.Subnets {
				if subnet.DefaultForAz != nil && *subnet.DefaultForAz {
					defaultSubnets = append(defaultSubnets, subnet)
					continue
				}
				subnets = append(subnets, subnet)
			}
			return !lastPage
		},
	)

	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	for _, item := range subnets {
		res := *item
		s.subnetRunner.Run(func() (cty.Value, error) {
			return s.readSubnet(res)
		})
	}

	for _, item := range defaultSubnets {
		res := *item
		s.defaultSubnetRunner.Run(func() (cty.Value, error) {
			return s.readSubnet(res)
		})
	}

	// Retrieve results from terraform provider
	defaultSubnetResources, err := s.defaultSubnetRunner.Wait()
	if err != nil {
		return nil, err
	}
	subnetResources, err := s.subnetRunner.Wait()
	if err != nil {
		return nil, err
	}

	// Deserialize
	deserializedDefaultSubnets, err := s.defaultSubnetDeserializer.Deserialize(defaultSubnetResources)
	if err != nil {
		return nil, err
	}
	deserializedSubnets, err := s.subnetDeserializer.Deserialize(subnetResources)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0, len(subnetResources)+len(deserializedDefaultSubnets))
	resources = append(resources, deserializedDefaultSubnets...)
	resources = append(resources, deserializedSubnets...)

	return resources, nil
}

func (s SubnetSupplier) readSubnet(subnet ec2.Subnet) (cty.Value, error) {
	var Ty resource.ResourceType = aws.AwsSubnetResourceType
	if subnet.DefaultForAz != nil && *subnet.DefaultForAz {
		Ty = aws.AwsDefaultSubnetResourceType
	}
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *subnet.SubnetId,
		Ty: Ty,
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}

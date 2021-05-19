package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

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
	client                    repository.EC2Repository
	defaultSubnetRunner       *terraform.ParallelResourceReader
	subnetRunner              *terraform.ParallelResourceReader
}

func NewSubnetSupplier(provider *AWSTerraformProvider) *SubnetSupplier {
	return &SubnetSupplier{
		provider,
		awsdeserializer.NewDefaultSubnetDeserializer(),
		awsdeserializer.NewSubnetDeserializer(),
		repository.NewEC2Repository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *SubnetSupplier) Resources() ([]resource.Resource, error) {
	subnets, defaultSubnets, err := s.client.ListAllSubnets()

	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsSubnetResourceType)
	}

	for _, item := range subnets {
		res := *item
		s.subnetRunner.Run(func() (cty.Value, error) {
			return s.readSubnet(res)
		})
	}

	subnetResources, err := s.subnetRunner.Wait()
	if err != nil {
		return nil, err
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

func (s *SubnetSupplier) readSubnet(subnet ec2.Subnet) (cty.Value, error) {
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

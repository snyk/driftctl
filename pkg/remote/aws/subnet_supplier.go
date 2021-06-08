package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/zclconf/go-cty/cty"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
)

type SubnetSupplier struct {
	reader              terraform.ResourceReader
	deserializer        *resource.Deserializer
	repo                repository.EC2Repository
	defaultSubnetRunner *terraform.ParallelResourceReader
	subnetRunner        *terraform.ParallelResourceReader
}

func NewSubnetSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.EC2Repository) *SubnetSupplier {
	return &SubnetSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *SubnetSupplier) SuppliedType() resource.ResourceType {
	return aws.AwsSubnetResourceType
}

func (s *SubnetSupplier) Resources() ([]resource.Resource, error) {
	subnets, defaultSubnets, err := s.repo.ListAllSubnets()

	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
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
	deserializedDefaultSubnets, err := s.deserializer.Deserialize(aws.AwsDefaultSubnetResourceType, defaultSubnetResources)
	if err != nil {
		return nil, err
	}
	deserializedSubnets, err := s.deserializer.Deserialize(s.SuppliedType(), subnetResources)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0, len(subnetResources)+len(deserializedDefaultSubnets))
	resources = append(resources, deserializedDefaultSubnets...)
	resources = append(resources, deserializedSubnets...)

	return resources, nil
}

func (s *SubnetSupplier) readSubnet(subnet ec2.Subnet) (cty.Value, error) {
	var Ty resource.ResourceType = s.SuppliedType()
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

package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/zclconf/go-cty/cty"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
)

type VPCSupplier struct {
	reader           terraform.ResourceReader
	deserializer     *resource.Deserializer
	client           ec2iface.EC2API
	defaultVPCRunner *terraform.ParallelResourceReader
	vpcRunner        *terraform.ParallelResourceReader
}

func NewVPCSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer) *VPCSupplier {
	return &VPCSupplier{
		provider,
		deserializer,
		ec2.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *VPCSupplier) Resources() ([]resource.Resource, error) {
	input := ec2.DescribeVpcsInput{}
	var VPCs []*ec2.Vpc
	var defaultVPCs []*ec2.Vpc
	err := s.client.DescribeVpcsPages(&input,
		func(resp *ec2.DescribeVpcsOutput, lastPage bool) bool {
			for _, vpc := range resp.Vpcs {
				if vpc.IsDefault != nil && *vpc.IsDefault {
					defaultVPCs = append(defaultVPCs, vpc)
					continue
				}
				VPCs = append(VPCs, vpc)
			}
			return !lastPage
		},
	)

	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsVpcResourceType)
	}

	for _, item := range VPCs {
		res := *item
		s.vpcRunner.Run(func() (cty.Value, error) {
			return s.readVPC(res)
		})
	}

	VPCResources, err := s.vpcRunner.Wait()
	if err != nil {
		return nil, err
	}

	for _, item := range defaultVPCs {
		res := *item
		s.defaultVPCRunner.Run(func() (cty.Value, error) {
			return s.readVPC(res)
		})
	}

	// Retrieve results from terraform provider
	defaultVPCResources, err := s.defaultVPCRunner.Wait()
	if err != nil {
		return nil, err
	}

	// Deserialize
	deserializedDefaultVPCs, err := s.deserializer.Deserialize(aws.AwsDefaultVpcResourceType, defaultVPCResources)
	if err != nil {
		return nil, err
	}
	deserializedVPCs, err := s.deserializer.Deserialize(aws.AwsVpcResourceType, VPCResources)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0, len(VPCResources)+len(deserializedDefaultVPCs))
	resources = append(resources, deserializedDefaultVPCs...)
	resources = append(resources, deserializedVPCs...)

	return resources, nil
}

func (s *VPCSupplier) readVPC(vpc ec2.Vpc) (cty.Value, error) {
	var Ty resource.ResourceType = aws.AwsVpcResourceType
	if vpc.IsDefault != nil && *vpc.IsDefault {
		Ty = aws.AwsDefaultVpcResourceType
	}
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *vpc.VpcId,
		Ty: Ty,
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}

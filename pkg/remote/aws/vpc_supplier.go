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

type VPCSupplier struct {
	reader                 terraform.ResourceReader
	defaultVPCDeserializer deserializer.CTYDeserializer
	vpcDeserializer        deserializer.CTYDeserializer
	client                 repository.EC2Repository
	defaultVPCRunner       *terraform.ParallelResourceReader
	vpcRunner              *terraform.ParallelResourceReader
}

func NewVPCSupplier(provider *AWSTerraformProvider) *VPCSupplier {
	return &VPCSupplier{
		provider,
		awsdeserializer.NewDefaultVPCDeserializer(),
		awsdeserializer.NewVPCDeserializer(),
		repository.NewEC2Repository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *VPCSupplier) Resources() ([]resource.Resource, error) {
	VPCs, defaultVPCs, err := s.client.ListAllVPCs()

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
	deserializedDefaultVPCs, err := s.defaultVPCDeserializer.Deserialize(defaultVPCResources)
	if err != nil {
		return nil, err
	}
	deserializedVPCs, err := s.vpcDeserializer.Deserialize(VPCResources)
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

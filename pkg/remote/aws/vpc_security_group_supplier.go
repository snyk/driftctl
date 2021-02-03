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

type VPCSecurityGroupSupplier struct {
	reader                           terraform.ResourceReader
	defaultSecurityGroupDeserializer deserializer.CTYDeserializer
	securityGroupDeserializer        deserializer.CTYDeserializer
	client                           ec2iface.EC2API
	defaultSecurityGroupRunner       *terraform.ParallelResourceReader
	securityGroupRunner              *terraform.ParallelResourceReader
}

func NewVPCSecurityGroupSupplier(provider *TerraformProvider) *VPCSecurityGroupSupplier {
	return &VPCSecurityGroupSupplier{
		provider,
		awsdeserializer.NewDefaultSecurityGroupDeserializer(),
		awsdeserializer.NewVPCSecurityGroupDeserializer(),
		ec2.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s VPCSecurityGroupSupplier) Resources() ([]resource.Resource, error) {
	securityGroups, defaultSecurityGroups, err := listSecurityGroups(s.client)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsSecurityGroupResourceType)
	}

	for _, item := range securityGroups {
		res := *item
		s.securityGroupRunner.Run(func() (cty.Value, error) {
			return s.readSecurityGroup(res)
		})
	}
	securityGroupResources, err := s.securityGroupRunner.Wait()
	if err != nil {
		return nil, err
	}

	for _, item := range defaultSecurityGroups {
		res := *item
		s.defaultSecurityGroupRunner.Run(func() (cty.Value, error) {
			return s.readSecurityGroup(res)
		})
	}
	defaultSecurityGroupResources, err := s.defaultSecurityGroupRunner.Wait()
	if err != nil {
		return nil, err
	}

	// Deserialize
	deserializedDefaultSecurityGroups, err := s.defaultSecurityGroupDeserializer.Deserialize(defaultSecurityGroupResources)
	if err != nil {
		return nil, err
	}
	deserializedSecurityGroups, err := s.securityGroupDeserializer.Deserialize(securityGroupResources)
	if err != nil {
		return nil, err
	}

	resources := make([]resource.Resource, 0, len(securityGroupResources)+len(defaultSecurityGroupResources))
	resources = append(resources, deserializedDefaultSecurityGroups...)
	resources = append(resources, deserializedSecurityGroups...)

	return resources, nil
}

func (s VPCSecurityGroupSupplier) readSecurityGroup(securityGroup ec2.SecurityGroup) (cty.Value, error) {
	var Ty resource.ResourceType = resourceaws.AwsSecurityGroupResourceType
	if isDefaultSecurityGroup(securityGroup) {
		Ty = resourceaws.AwsDefaultSecurityGroupResourceType
	}
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: aws.StringValue(securityGroup.GroupId),
		Ty: Ty,
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}

func listSecurityGroups(client ec2iface.EC2API) ([]*ec2.SecurityGroup, []*ec2.SecurityGroup, error) {
	var securityGroups []*ec2.SecurityGroup
	var defaultSecurityGroups []*ec2.SecurityGroup
	input := &ec2.DescribeSecurityGroupsInput{}
	err := client.DescribeSecurityGroupsPages(input, func(res *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		for _, securityGroup := range res.SecurityGroups {
			if isDefaultSecurityGroup(*securityGroup) {
				defaultSecurityGroups = append(defaultSecurityGroups, securityGroup)
				continue
			}
			securityGroups = append(securityGroups, securityGroup)
		}
		return !lastPage
	})
	if err != nil {
		return nil, nil, err
	}
	return securityGroups, defaultSecurityGroups, nil
}

// Return true if the security group is considered as a default one
func isDefaultSecurityGroup(securityGroup ec2.SecurityGroup) bool {
	return securityGroup.GroupName != nil && *securityGroup.GroupName == "default"
}

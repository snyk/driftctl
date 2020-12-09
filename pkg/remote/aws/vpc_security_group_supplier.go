package aws

import (
	"github.com/cloudskiff/driftctl/pkg"
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
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       ec2iface.EC2API
	runner       *terraform.ParallelResourceReader
}

func NewVPCSecurityGroupSupplier(runner *pkg.ParallelRunner, client ec2iface.EC2API) *VPCSecurityGroupSupplier {
	return &VPCSecurityGroupSupplier{terraform.Provider(terraform.AWS), awsdeserializer.NewVPCSecurityGroupDeserializer(), client, terraform.NewParallelResourceReader(runner)}
}

func (s VPCSecurityGroupSupplier) Resources() ([]resource.Resource, error) {
	securityGroups, err := listSecurityGroups(s.client)
	if err != nil {
		return nil, err
	}
	results := make([]cty.Value, 0)
	if len(securityGroups) > 0 {
		for _, securityGroup := range securityGroups {
			sg := *securityGroup
			s.runner.Run(func() (cty.Value, error) {
				return s.readSecurityGroup(sg)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(results)
}

func (s VPCSecurityGroupSupplier) readSecurityGroup(securityGroup ec2.SecurityGroup) (cty.Value, error) {
	id := aws.StringValue(securityGroup.GroupId)
	resSg, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: resourceaws.AwsSecurityGroupResourceType,
		ID: id,
	})
	if err != nil {
		logrus.Warnf("Error reading security group %s: %+v", id, err)
		return cty.NilVal, err
	}
	return *resSg, nil
}

func listSecurityGroups(client ec2iface.EC2API) ([]*ec2.SecurityGroup, error) {
	var securityGroups []*ec2.SecurityGroup
	input := &ec2.DescribeSecurityGroupsInput{}
	err := client.DescribeSecurityGroupsPages(input, func(res *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		securityGroups = append(securityGroups, res.SecurityGroups...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return securityGroups, nil
}

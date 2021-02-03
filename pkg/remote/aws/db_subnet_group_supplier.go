package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/zclconf/go-cty/cty"

	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/aws/aws-sdk-go/service/rds/rdsiface"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
)

type DBSubnetGroupSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       rdsiface.RDSAPI
	runner       *terraform.ParallelResourceReader
}

func NewDBSubnetGroupSupplier(provider *TerraformProvider) *DBSubnetGroupSupplier {
	return &DBSubnetGroupSupplier{
		provider,
		awsdeserializer.NewDBSubnetGroupDeserializer(),
		rds.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s DBSubnetGroupSupplier) Resources() ([]resource.Resource, error) {

	input := rds.DescribeDBSubnetGroupsInput{}
	var subnetGroups []*rds.DBSubnetGroup
	err := s.client.DescribeDBSubnetGroupsPages(&input,
		func(resp *rds.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
			subnetGroups = append(subnetGroups, resp.DBSubnetGroups...)
			return !lastPage
		},
	)

	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsDbSubnetGroupResourceType)
	}

	for _, subnetGroup := range subnetGroups {
		sub := *subnetGroup
		s.runner.Run(func() (cty.Value, error) {
			return s.readSubnetGroup(sub)
		})
	}
	ctyValues, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}
	return s.deserializer.Deserialize(ctyValues)
}

func (s DBSubnetGroupSupplier) readSubnetGroup(subnetGroup rds.DBSubnetGroup) (cty.Value, error) {
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *subnetGroup.DBSubnetGroupName,
		Ty: aws.AwsDbSubnetGroupResourceType,
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}

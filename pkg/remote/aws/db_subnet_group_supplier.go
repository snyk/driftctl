package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/zclconf/go-cty/cty"

	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
)

type DBSubnetGroupSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.RDSRepository
	runner       *terraform.ParallelResourceReader
}

func NewDBSubnetGroupSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.RDSRepository) *DBSubnetGroupSupplier {
	return &DBSubnetGroupSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *DBSubnetGroupSupplier) SuppliedType() resource.ResourceType {
	return aws.AwsDbSubnetGroupResourceType
}

func (s *DBSubnetGroupSupplier) Resources() ([]resource.Resource, error) {

	subnetGroups, err := s.client.ListAllDbSubnetGroups()

	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
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
	return s.deserializer.Deserialize(s.SuppliedType(), ctyValues)
}

func (s *DBSubnetGroupSupplier) readSubnetGroup(subnetGroup rds.DBSubnetGroup) (cty.Value, error) {
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *subnetGroup.DBSubnetGroupName,
		Ty: s.SuppliedType(),
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}

package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type DynamoDBTableSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repository   repository.DynamoDBRepository
	runner       *terraform.ParallelResourceReader
}

func NewDynamoDBTableSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.DynamoDBRepository) *DynamoDBTableSupplier {
	return &DynamoDBTableSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *DynamoDBTableSupplier) SuppliedType() resource.ResourceType {
	return aws.AwsDynamodbTableResourceType
}

func (s *DynamoDBTableSupplier) Resources() ([]resource.Resource, error) {
	tables, err := s.repository.ListAllTables()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}

	for _, table := range tables {
		table := table
		s.runner.Run(func() (cty.Value, error) {
			return s.readTable(table)
		})
	}

	retrieve, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(s.SuppliedType(), retrieve)
}

func (s *DynamoDBTableSupplier) readTable(tableName *string) (cty.Value, error) {
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *tableName,
		Ty: s.SuppliedType(),
		Attributes: map[string]string{
			"table_name": *tableName,
		},
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}

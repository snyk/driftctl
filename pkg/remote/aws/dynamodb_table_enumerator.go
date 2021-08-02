package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type DynamoDBTableEnumerator struct {
	repository repository.DynamoDBRepository
	factory    resource.ResourceFactory
}

func NewDynamoDBTableEnumerator(repository repository.DynamoDBRepository, factory resource.ResourceFactory) *DynamoDBTableEnumerator {
	return &DynamoDBTableEnumerator{
		repository,
		factory,
	}
}

func (e *DynamoDBTableEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsDynamodbTableResourceType
}

func (e *DynamoDBTableEnumerator) Enumerate() ([]resource.Resource, error) {
	tables, err := e.repository.ListAllTables()
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(tables))

	for _, table := range tables {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*table,
				map[string]interface{}{},
			),
		)
	}

	return results, nil
}

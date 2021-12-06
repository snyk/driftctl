package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
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

func (e *DynamoDBTableEnumerator) Enumerate() ([]*resource.Resource, error) {
	tables, err := e.repository.ListAllTables()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(tables))

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

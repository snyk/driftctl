package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type RDSDBInstanceEnumerator struct {
	repository repository.RDSRepository
	factory    resource.ResourceFactory
}

func NewRDSDBInstanceEnumerator(repo repository.RDSRepository, factory resource.ResourceFactory) *RDSDBInstanceEnumerator {
	return &RDSDBInstanceEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *RDSDBInstanceEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsDbInstanceResourceType
}

func (e *RDSDBInstanceEnumerator) Enumerate() ([]*resource.Resource, error) {
	instances, err := e.repository.ListAllDBInstances()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(instances))

	for _, instance := range instances {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*instance.DBInstanceIdentifier,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

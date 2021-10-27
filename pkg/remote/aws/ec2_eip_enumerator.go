package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type EC2EipEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2EipEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2EipEnumerator {
	return &EC2EipEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2EipEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsEipResourceType
}

func (e *EC2EipEnumerator) Enumerate() ([]*resource.Resource, error) {
	addresses, err := e.repository.ListAllAddresses()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(addresses))

	for _, address := range addresses {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*address.AllocationId,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

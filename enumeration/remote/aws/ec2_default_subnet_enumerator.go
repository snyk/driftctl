package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type EC2DefaultSubnetEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2DefaultSubnetEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2DefaultSubnetEnumerator {
	return &EC2DefaultSubnetEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2DefaultSubnetEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsDefaultSubnetResourceType
}

func (e *EC2DefaultSubnetEnumerator) Enumerate() ([]*resource.Resource, error) {
	_, defaultSubnets, err := e.repository.ListAllSubnets()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(defaultSubnets))

	for _, subnet := range defaultSubnets {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*subnet.SubnetId,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

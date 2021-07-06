package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type EC2NatGatewayEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2NatGatewayEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2NatGatewayEnumerator {
	return &EC2NatGatewayEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2NatGatewayEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsNatGatewayResourceType
}

func (e *EC2NatGatewayEnumerator) Enumerate() ([]resource.Resource, error) {
	natGateways, err := e.repository.ListAllNatGateways()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(natGateways))

	for _, natGateway := range natGateways {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*natGateway.NatGatewayId,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

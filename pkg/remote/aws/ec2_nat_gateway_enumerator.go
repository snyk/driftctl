package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
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

func (e *EC2NatGatewayEnumerator) Enumerate() ([]*resource.Resource, error) {
	natGateways, err := e.repository.ListAllNatGateways()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(natGateways))

	for _, natGateway := range natGateways {

		attrs := map[string]interface{}{}
		if len(natGateway.NatGatewayAddresses) > 0 {
			if allocId := natGateway.NatGatewayAddresses[0].AllocationId; allocId != nil {
				attrs["allocation_id"] = *allocId
			}
		}

		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*natGateway.NatGatewayId,
				attrs,
			),
		)
	}

	return results, err
}

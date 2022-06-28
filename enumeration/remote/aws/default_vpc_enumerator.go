package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource/aws"

	"github.com/snyk/driftctl/enumeration/resource"
)

type DefaultVPCEnumerator struct {
	repo    repository.EC2Repository
	factory resource.ResourceFactory
}

func NewDefaultVPCEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *DefaultVPCEnumerator {
	return &DefaultVPCEnumerator{
		repo,
		factory,
	}
}

func (e *DefaultVPCEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsDefaultVpcResourceType
}

func (e *DefaultVPCEnumerator) Enumerate() ([]*resource.Resource, error) {
	_, defaultVPCs, err := e.repo.ListAllVPCs()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(defaultVPCs))

	for _, item := range defaultVPCs {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*item.VpcId,
				map[string]interface{}{},
			),
		)
	}

	return results, nil
}

package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/resource"
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

func (e *DefaultVPCEnumerator) Enumerate() ([]resource.Resource, error) {
	_, defaultVPCs, err := e.repo.ListAllVPCs()
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, aws.AwsDefaultVpcResourceType)
	}

	results := make([]resource.Resource, 0, len(defaultVPCs))

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

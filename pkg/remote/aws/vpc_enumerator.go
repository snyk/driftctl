package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

type VPCEnumerator struct {
	repo    repository.EC2Repository
	factory resource.ResourceFactory
}

func NewVPCEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *VPCEnumerator {
	return &VPCEnumerator{
		repo,
		factory,
	}
}

func (e *VPCEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsVpcResourceType
}

func (e *VPCEnumerator) Enumerate() ([]resource.Resource, error) {
	VPCs, _, err := e.repo.ListAllVPCs()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, 0, len(VPCs))

	for _, item := range VPCs {
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

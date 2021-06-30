package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
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
	VPCs, defaultVPCs, err := e.repo.ListAllVPCs()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsVpcResourceType)
	}

	results := make([]resource.Resource, 0, len(VPCs)+len(defaultVPCs))

	for _, item := range VPCs {
		results = append(
			results,
			e.readVPC(*item),
		)
	}

	for _, item := range defaultVPCs {
		results = append(
			results,
			e.readVPC(*item),
		)
	}

	return results, nil
}

func (e *VPCEnumerator) readVPC(vpc ec2.Vpc) resource.Resource {
	var Ty = e.SupportedType()
	if vpc.IsDefault != nil && *vpc.IsDefault {
		Ty = aws.AwsDefaultVpcResourceType
	}

	res := e.factory.CreateAbstractResource(
		string(Ty),
		*vpc.VpcId,
		map[string]interface{}{},
	)

	return res
}

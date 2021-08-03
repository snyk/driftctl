package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws"
)

type VPCDefaultSecurityGroupEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewVPCDefaultSecurityGroupEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *VPCDefaultSecurityGroupEnumerator {
	return &VPCDefaultSecurityGroupEnumerator{
		repo,
		factory,
	}
}

func (e *VPCDefaultSecurityGroupEnumerator) SupportedType() resource.ResourceType {
	return resourceaws.AwsDefaultSecurityGroupResourceType
}

func (e *VPCDefaultSecurityGroupEnumerator) Enumerate() ([]resource.Resource, error) {
	_, defaultSecurityGroups, err := e.repository.ListAllSecurityGroups()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, 0, len(defaultSecurityGroups))

	for _, item := range defaultSecurityGroups {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				aws.StringValue(item.GroupId),
				map[string]interface{}{},
			),
		)
	}

	return results, nil
}

package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type EC2DefaultNetworkACLEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2DefaultNetworkACLEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2DefaultNetworkACLEnumerator {
	return &EC2DefaultNetworkACLEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2DefaultNetworkACLEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsDefaultNetworkACLResourceType
}

func (e *EC2DefaultNetworkACLEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.ListAllNetworkACLs()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(resources))

	for _, res := range resources {
		// Do not handle non-default network acl since it is a dedicated resource
		if !*res.IsDefault {
			continue
		}
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*res.NetworkAclId,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

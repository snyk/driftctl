package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type EC2EipAssociationEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2EipAssociationEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2EipAssociationEnumerator {
	return &EC2EipAssociationEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2EipAssociationEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsEipAssociationResourceType
}

func (e *EC2EipAssociationEnumerator) Enumerate() ([]resource.Resource, error) {
	associationIds, err := e.repository.ListAllAddressesAssociation()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(associationIds))

	for _, associationId := range associationIds {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				associationId,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

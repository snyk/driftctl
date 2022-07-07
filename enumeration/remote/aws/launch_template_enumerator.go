package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type LaunchTemplateEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewLaunchTemplateEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *LaunchTemplateEnumerator {
	return &LaunchTemplateEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *LaunchTemplateEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsLaunchTemplateResourceType
}

func (e *LaunchTemplateEnumerator) Enumerate() ([]*resource.Resource, error) {
	templates, err := e.repository.DescribeLaunchTemplates()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(templates))

	for _, tmpl := range templates {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*tmpl.LaunchTemplateId,
				map[string]interface{}{},
			),
		)
	}

	return results, nil
}

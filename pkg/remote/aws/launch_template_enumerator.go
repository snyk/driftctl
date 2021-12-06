package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
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
		launchTemplateVersions, err := e.repository.DescribeLaunchTemplateVersions(tmpl)
		if err != nil {
			return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
		}
		ltData := launchTemplateVersions[0].LaunchTemplateData

		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*tmpl.LaunchTemplateId,
				map[string]interface{}{
					"credit_specification": getCreditSpecification(ltData.CreditSpecification),
				},
			),
		)
	}

	return results, nil
}

func getCreditSpecification(cs *ec2.CreditSpecification) []interface{} {
	spec := make([]interface{}, 0)
	if cs != nil {
		spec = append(spec, map[string]interface{}{
			"cpu_credits": *cs.CpuCredits,
		})
	}
	return spec
}

package aws

import (
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type Route53HealthCheckEnumerator struct {
	repository repository.Route53Repository
	factory    resource.ResourceFactory
}

func NewRoute53HealthCheckEnumerator(repo repository.Route53Repository, factory resource.ResourceFactory) *Route53HealthCheckEnumerator {
	return &Route53HealthCheckEnumerator{
		repo,
		factory,
	}
}

func (e *Route53HealthCheckEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsRoute53HealthCheckResourceType
}

func (e *Route53HealthCheckEnumerator) Enumerate() ([]*resource.Resource, error) {
	healthChecks, err := e.repository.ListAllHealthChecks()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(healthChecks))

	for _, healthCheck := range healthChecks {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*healthCheck.Id,
				map[string]interface{}{},
			),
		)
	}

	return results, nil
}

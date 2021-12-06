package google

import (
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
)

type GoogleComputeHealthCheckEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleComputeHealthCheckEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleComputeHealthCheckEnumerator {
	return &GoogleComputeHealthCheckEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleComputeHealthCheckEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleComputeHealthCheckResourceType
}

func (e *GoogleComputeHealthCheckEnumerator) Enumerate() ([]*resource.Resource, error) {
	checks, err := e.repository.SearchAllHealthChecks()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(checks))
	for _, res := range checks {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				trimResourceName(res.GetName()),
				map[string]interface{}{
					"name": res.GetDisplayName(),
				},
			),
		)
	}

	return results, err
}

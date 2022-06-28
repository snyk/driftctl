package google

import (
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/remote/google/repository"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
)

type GoogleComputeNodeGroupEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleComputeNodeGroupEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleComputeNodeGroupEnumerator {
	return &GoogleComputeNodeGroupEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleComputeNodeGroupEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleComputeNodeGroupResourceType
}

func (e *GoogleComputeNodeGroupEnumerator) Enumerate() ([]*resource.Resource, error) {
	nodeGroups, err := e.repository.SearchAllNodeGroups()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(nodeGroups))
	for _, res := range nodeGroups {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				trimResourceName(res.GetName()),
				map[string]interface{}{
					"name": res.GetName(),
				},
			),
		)
	}

	return results, err
}

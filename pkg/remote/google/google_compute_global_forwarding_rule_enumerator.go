package google

import (
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
)

type GoogleComputeGlobalForwardingRuleEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleComputeGlobalForwardingRuleEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleComputeGlobalForwardingRuleEnumerator {
	return &GoogleComputeGlobalForwardingRuleEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleComputeGlobalForwardingRuleEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleComputeGlobalForwardingRuleResourceType
}

func (e *GoogleComputeGlobalForwardingRuleEnumerator) Enumerate() ([]*resource.Resource, error) {
	nodeGroups, err := e.repository.SearchAllGlobalForwardingRules() //Change the name
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
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

package google

import (
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/remote/google/repository"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
)

type GoogleComputeForwardingRuleEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleComputeForwardingRuleEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleComputeForwardingRuleEnumerator {
	return &GoogleComputeForwardingRuleEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleComputeForwardingRuleEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleComputeForwardingRuleResourceType
}

func (e *GoogleComputeForwardingRuleEnumerator) Enumerate() ([]*resource.Resource, error) {
	forwardingRules, err := e.repository.SearchAllForwardingRules()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(forwardingRules))
	for _, res := range forwardingRules {
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

package google

import (
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/remote/google/repository"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
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
	globalForwardingRules, err := e.repository.SearchAllGlobalForwardingRules()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(globalForwardingRules))

	for _, res := range globalForwardingRules {
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

package google

import (
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/remote/google/repository"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
)

type GoogleComputeSslCertificateRuleEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleComputeSslCertificateRuleEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleComputeSslCertificateRuleEnumerator {
	return &GoogleComputeSslCertificateRuleEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleComputeSslCertificateRuleEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleComputeSslCertificateRuleResourceType
}

func (e *GoogleComputeSslCertificateRuleEnumerator) Enumerate() ([]*resource.Resource, error) {
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

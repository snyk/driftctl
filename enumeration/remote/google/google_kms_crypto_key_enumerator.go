package google

import (
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/remote/google/repository"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
)

type GoogleKmsCryptoKeyEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleKmsCryptoKeyEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleKmsCryptoKeyEnumerator {
	return &GoogleKmsCryptoKeyEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleKmsCryptoKeyEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleKmsCryptoKeyResourceType
}

func (e *GoogleKmsCryptoKeyEnumerator) Enumerate() ([]*resource.Resource, error) {
	kmsCryptoKeys, err := e.repository.SearchAllKmsCryptoKeys()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(kmsCryptoKeys))
	for _, res := range kmsCryptoKeys {
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

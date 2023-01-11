package google

import (
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/remote/google/repository"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
)

type GoogleComputeSslCertificateEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleComputeSslCertificateEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleComputeSslCertificateEnumerator {
	return &GoogleComputeSslCertificateEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleComputeSslCertificateEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleComputeSslCertificateResourceType
}

func (e *GoogleComputeSslCertificateEnumerator) Enumerate() ([]*resource.Resource, error) {
	sslCertificates, err := e.repository.SearchAllSslCertificates()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(sslCertificates))
	for _, res := range sslCertificates {
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

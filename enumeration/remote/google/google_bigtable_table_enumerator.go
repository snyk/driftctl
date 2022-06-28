package google

import (
	"github.com/sirupsen/logrus"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/remote/google/repository"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/google"
)

type GoogleBigtableTableEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleBigtableTableEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleBigtableTableEnumerator {
	return &GoogleBigtableTableEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleBigtableTableEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleBigtableTableResourceType
}

func (e *GoogleBigtableTableEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.SearchAllBigtableTables()

	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(resources))

	for _, res := range resources {
		name, exist := res.GetResource().GetData().GetFields()["name"]
		if !exist || name.GetStringValue() == "" {
			logrus.WithField("name", res.GetName()).Warn("Unable to retrieve resource name")
			continue
		}
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				name.GetStringValue(),
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

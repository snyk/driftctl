package google

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/remote/google/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/google"
	"github.com/sirupsen/logrus"
)

type GoogleBigTableInstanceEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleBigTableInstanceEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleBigTableInstanceEnumerator {
	return &GoogleBigTableInstanceEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleBigTableInstanceEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleBigTableInstanceResourceType
}

func (e *GoogleBigTableInstanceEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.SearchAllBigtableInstances()

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

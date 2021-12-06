package google

import (
	"strings"

	"github.com/sirupsen/logrus"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
)

type GoogleCloudRunServiceEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleCloudRunServiceEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleCloudRunServiceEnumerator {
	return &GoogleCloudRunServiceEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleCloudRunServiceEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleCloudRunServiceResourceType
}

func (e *GoogleCloudRunServiceEnumerator) Enumerate() ([]*resource.Resource, error) {
	subnets, err := e.repository.SearchAllCloudRunServices()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(subnets))

	for _, res := range subnets {
		splittedName := strings.Split(res.GetName(), "/")
		if len(splittedName) != 9 {
			logrus.WithField("name", res.GetName()).Error("Unable to decode project from resource name")
			continue
		}
		project := splittedName[4]
		id := strings.Join([]string{
			"locations", res.GetLocation(),
			"namespaces", project,
			"services", res.GetDisplayName(),
		}, "/")
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				id,
				map[string]interface{}{
					"name": res.GetDisplayName(),
				},
			),
		)
	}

	return results, err
}

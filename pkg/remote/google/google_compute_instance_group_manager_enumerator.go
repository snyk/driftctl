package google

import (
	"strings"

	"github.com/sirupsen/logrus"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
)

type GoogleComputeInstanceGroupManagerEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleComputeInstanceGroupManagerEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleComputeInstanceGroupManagerEnumerator {
	return &GoogleComputeInstanceGroupManagerEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleComputeInstanceGroupManagerEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleComputeInstanceGroupManagerResourceType
}

func (e *GoogleComputeInstanceGroupManagerEnumerator) Enumerate() ([]*resource.Resource, error) {
	items, err := e.repository.SearchAllInstanceGroupManagers()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(items))
	for _, res := range items {
		splittedName := strings.Split(res.GetName(), "/")
		if len(splittedName) != 9 {
			logrus.WithField("name", res.GetName()).Error("Unable to decode project from instance group name")
			continue
		}
		name := splittedName[8]
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				trimResourceName(res.GetName()),
				map[string]interface{}{
					"name": name,
				},
			),
		)
	}

	return results, err
}

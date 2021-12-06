package google

import (
	"strings"

	"github.com/sirupsen/logrus"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
)

type GoogleComputeInstanceGroupEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleComputeInstanceGroupEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleComputeInstanceGroupEnumerator {
	return &GoogleComputeInstanceGroupEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleComputeInstanceGroupEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleComputeInstanceGroupResourceType
}

func (e *GoogleComputeInstanceGroupEnumerator) Enumerate() ([]*resource.Resource, error) {
	groups, err := e.repository.SearchAllInstanceGroups()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(groups))
	for _, res := range groups {
		splittedName := strings.Split(res.GetName(), "/")
		if len(splittedName) != 9 {
			logrus.WithField("name", res.GetName()).Error("Unable to decode project from instance group name")
			continue
		}
		project := splittedName[4]
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				trimResourceName(res.GetName()),
				map[string]interface{}{
					"name":     res.GetDisplayName(),
					"project":  project,
					"location": res.GetLocation(),
				},
			),
		)
	}

	return results, err
}

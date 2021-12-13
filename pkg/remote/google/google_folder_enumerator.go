package google

import (
	"strings"

	"github.com/sirupsen/logrus"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
)

type GoogleFolderEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleFolderEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleFolderEnumerator {
	return &GoogleFolderEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleFolderEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleFolderResourceType
}

func (e *GoogleFolderEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.SearchAllFolders()

	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(resources))

	for _, res := range resources {
		id := trimResourceName(res.Name)
		splittedId := strings.Split(id, "folders/")
		if len(splittedId) != 2 {
			logrus.WithField("id", res.Name).Warn("Cannot parse google_folder ID")
			continue
		}
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				id,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

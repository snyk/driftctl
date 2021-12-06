package google

import (
	"strings"

	"github.com/sirupsen/logrus"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
)

type GoogleDNSManagedZoneEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleDNSManagedZoneEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleDNSManagedZoneEnumerator {
	return &GoogleDNSManagedZoneEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleDNSManagedZoneEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleDNSManagedZoneResourceType
}

func (e *GoogleDNSManagedZoneEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.SearchAllDNSManagedZones()

	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(resources))

	// We should have ID = "projects/cloudskiff-dev-elie/managedZones/example-zone"
	// We have projects/cloudskiff-dev-elie/managedZones/2435093289230056557
	for _, res := range resources {
		id := trimResourceName(res.Name)
		splittedId := strings.Split(id, "/managedZones/")
		if len(splittedId) != 2 {
			logrus.WithField("id", res.Name).Warn("Cannot parse google_dns_managed_zone ID")
			continue
		}
		id = strings.Join([]string{splittedId[0], "managedZones", res.DisplayName}, "/")
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

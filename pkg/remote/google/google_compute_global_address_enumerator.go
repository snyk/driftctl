package google

import (
	"github.com/sirupsen/logrus"
	remoteerror "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/remote/google/repository"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/google"
)

type GoogleComputeGlobalAddressEnumerator struct {
	repository repository.AssetRepository
	factory    resource.ResourceFactory
}

func NewGoogleComputeGlobalAddressEnumerator(repo repository.AssetRepository, factory resource.ResourceFactory) *GoogleComputeGlobalAddressEnumerator {
	return &GoogleComputeGlobalAddressEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *GoogleComputeGlobalAddressEnumerator) SupportedType() resource.ResourceType {
	return google.GoogleComputeGlobalAddressResourceType
}

func (e *GoogleComputeGlobalAddressEnumerator) Enumerate() ([]*resource.Resource, error) {
	resources, err := e.repository.SearchAllGlobalAddresses()

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
		address := ""
		if addr, exist := res.GetResource().GetData().GetFields()["address"]; exist {
			address = addr.GetStringValue()
		}
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				trimResourceName(res.GetName()),
				map[string]interface{}{
					"name":    name.GetStringValue(),
					"address": address,
				},
			),
		)
	}

	return results, err
}

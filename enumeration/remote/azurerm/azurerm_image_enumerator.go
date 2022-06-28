package azurerm

import (
	"github.com/snyk/driftctl/enumeration/remote/azurerm/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"strings"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/azurerm"
)

type AzurermImageEnumerator struct {
	repository repository.ComputeRepository
	factory    resource.ResourceFactory
}

func NewAzurermImageEnumerator(repo repository.ComputeRepository, factory resource.ResourceFactory) *AzurermImageEnumerator {
	return &AzurermImageEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *AzurermImageEnumerator) SupportedType() resource.ResourceType {
	return azurerm.AzureImageResourceType
}

func (e *AzurermImageEnumerator) Enumerate() ([]*resource.Resource, error) {
	images, err := e.repository.ListAllImages()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(images))

	for _, res := range images {
		r, err := azure.ParseResourceID(*res.ID)
		if err != nil {
			logrus.WithFields(map[string]interface{}{
				"id":   *res.ID,
				"type": string(e.SupportedType()),
			}).Error("Failed to parse Azure resource ID")
			continue
		}

		// Here we turn the resource group into lowercase because for some reason the API returns it in uppercase.
		resourceId := strings.Replace(*res.ID, r.ResourceGroup, strings.ToLower(r.ResourceGroup), 1)

		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				resourceId,
				map[string]interface{}{
					"name": *res.Name,
				},
			),
		)

	}

	return results, err
}

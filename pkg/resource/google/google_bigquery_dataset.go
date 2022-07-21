package google

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const GoogleBigqueryDatasetResourceType = "google_bigquery_dataset"

func initGoogleBigqueryDatasetMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleBigqueryDatasetResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"name": *res.Attrs.GetString("friendly_name"),
		}
	})
}

package google

import "github.com/snyk/driftctl/pkg/resource"

const GoogleBigqueryTableResourceType = "google_bigquery_table"

func initGoogleBigqueryTableMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetHumanReadableAttributesFunc(GoogleBigqueryTableResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"name": *res.Attrs.GetString("friendly_name"),
		}
	})
}

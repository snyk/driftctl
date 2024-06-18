package scaleway

import (
	"github.com/snyk/driftctl/pkg/resource"
)

func InitResourcesMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	initScalewayFunctionNamespace(resourceSchemaRepository)
}

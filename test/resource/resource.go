package resource

import (
	"github.com/hashicorp/terraform/providers"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/test/schemas"
)

func InitFakeSchemaRepository(provider, version string) resource.SchemaRepositoryInterface {
	repo := resource.NewSchemaRepository()
	schema := make(map[string]providers.Schema)
	if provider != "" {
		s, err := schemas.ReadTestSchema(provider, version)
		if err != nil {
			// TODO HANDLER ERROR PROPERLY
			panic(err)
		}
		schema = s
	}
	_ = repo.Init("Fake", "1.0.0", schema)
	return repo
}

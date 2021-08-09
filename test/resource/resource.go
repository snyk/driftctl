package resource

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/test/schemas"
	"github.com/hashicorp/terraform/providers"
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

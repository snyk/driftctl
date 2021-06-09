package terraform

import "github.com/hashicorp/terraform/providers"

// Representation of a TF Provider able to give it's schema and reade a resource
type TerraformProvider interface {
	SchemaSupplier
	ResourceReader
	Cleanup()
	TerraformProviderSchema() providers.GetSchemaResponse
}

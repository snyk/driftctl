package terraform

// Representation of a TF Provider able to give it's schema and reade a resource
type TerraformProvider interface {
	SchemaSupplier
	ResourceReader
	Cleanup()
	Name() string
	Version() string
}

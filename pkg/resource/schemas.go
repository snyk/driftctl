package resource

import "github.com/snyk/driftctl/enumeration/resource"

type SchemaRepositoryInterface interface {
	GetSchema(resourceType string) (*resource.Schema, bool)
	SetFlags(typ string, flags ...resource.Flags)
	UpdateSchema(typ string, schemasMutators map[string]func(attributeSchema *resource.AttributeSchema))
	SetNormalizeFunc(typ string, normalizeFunc func(res *resource.Resource))
	SetHumanReadableAttributesFunc(typ string, humanReadableAttributesFunc func(res *resource.Resource) map[string]string)
	SetDiscriminantFunc(string, func(*resource.Resource, *resource.Resource) bool)
}

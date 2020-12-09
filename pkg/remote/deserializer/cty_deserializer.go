package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"

	"github.com/zclconf/go-cty/cty"
)

// CTYDeserializer deserialize a resource.resource from cty return by ResourceReader or iac reader.
type CTYDeserializer interface {
	HandledType() resource.ResourceType
	Deserialize(values []cty.Value) ([]resource.Resource, error)
}

package enumeration

import (
	"time"

	"github.com/snyk/driftctl/enumeration/resource"
)

type EnumerateInput struct {
	ResourceTypes []string
}

type EnumerateOutput struct {
	// Resources is a map of resources by type. Every listed resource type will
	// have a key in the map. The value will be either nil or an empty slice if
	// no resources of that type were found.
	Resources map[string][]*resource.Resource

	// Timings is map of list durations by resource type. This aids understanding
	// which resource types took the most time to list.
	Timings map[string]time.Duration

	// Diagnostics contains messages and errors that arose during the list operation.
	// If the diagnostic is associated with a resource type, the ResourceType()
	// call will indicate which type. If associated with a resource, the Resource()
	// call will indicate which resource.
	Diagnostics Diagnostics
}

type Enumerator interface {
	Enumerate(*EnumerateInput) (*EnumerateOutput, error)
}

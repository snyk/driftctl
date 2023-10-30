package common

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

type Enumerator interface {
	SupportedType() resource.ResourceType
	Enumerate() ([]*resource.Resource, error)
}

type RemoteLibrary struct {
	enumerators []Enumerator
}

func NewRemoteLibrary() *RemoteLibrary {
	return &RemoteLibrary{
		make([]Enumerator, 0),
	}
}

func (r *RemoteLibrary) AddEnumerator(enumerator Enumerator) {
	r.enumerators = append(r.enumerators, enumerator)
}

func (r *RemoteLibrary) Enumerators() []Enumerator {
	return r.enumerators
}

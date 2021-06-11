package common

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

type Enumerator interface {
	SupportedType() resource.ResourceType
	Enumerate() ([]resource.Resource, error)
}

type DetailsFetcher interface {
	ReadDetails(resource.Resource) (resource.Resource, error)
}

type RemoteLibrary struct {
	enumerators     []Enumerator
	detailsFetchers map[resource.ResourceType]DetailsFetcher
}

func NewRemoteLibrary() *RemoteLibrary {
	return &RemoteLibrary{
		make([]Enumerator, 0),
		make(map[resource.ResourceType]DetailsFetcher),
	}
}

func (r *RemoteLibrary) AddEnumerator(enumerator Enumerator) {
	r.enumerators = append(r.enumerators, enumerator)
}

func (r *RemoteLibrary) Enumerators() []Enumerator {
	return r.enumerators
}

func (r *RemoteLibrary) AddDetailsFetcher(ty resource.ResourceType, detailFetcher DetailsFetcher) {
	r.detailsFetchers[ty] = detailFetcher
}

func (r *RemoteLibrary) GetDetailsFetcher(ty resource.ResourceType) DetailsFetcher {
	return r.detailsFetchers[ty]
}

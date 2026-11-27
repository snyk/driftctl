package repository

import (
	"github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/snyk/driftctl/enumeration/remote/cache"
)

type FunctionRepository interface {
	ListAllNamespaces() ([]*function.Namespace, error)
}

// We create an interface here (mainly for mocking purpose) because in scaleway-sdk-go, API is a struct and not an interface
type functionAPI interface {
	ListNamespaces(req *function.ListNamespacesRequest, opts ...scw.RequestOption) (*function.ListNamespacesResponse, error)
}

type functionRepository struct {
	client functionAPI
	cache  cache.Cache
}

func NewFunctionRepository(client *scw.Client, c cache.Cache) *functionRepository {

	api := function.NewAPI(client)
	return &functionRepository{
		api,
		c,
	}
}

func (r *functionRepository) ListAllNamespaces() ([]*function.Namespace, error) {
	if v := r.cache.Get("functionListAllNamespaces"); v != nil {
		return v.([]*function.Namespace), nil
	}

	req := &function.ListNamespacesRequest{}
	res, err := r.client.ListNamespaces(req)
	if err != nil {
		return nil, err
	}

	namespaces := res.Namespaces

	r.cache.Put("functionListAllNamespaces", namespaces)

	return namespaces, err
}

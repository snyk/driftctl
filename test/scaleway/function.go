package scaleway

import (
	"github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type FakeFunction interface {
	ListNamespaces(req *function.ListNamespacesRequest, opts ...scw.RequestOption) (*function.ListNamespacesResponse, error)
}

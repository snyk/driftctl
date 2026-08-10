package scaleway

import (
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/remote/scaleway/repository"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/scaleway"
)

type FunctionNamespaceEnumerator struct {
	repository repository.FunctionRepository
	factory    resource.ResourceFactory
}

func NewFunctionNamespaceEnumerator(repo repository.FunctionRepository, factory resource.ResourceFactory) *FunctionNamespaceEnumerator {
	return &FunctionNamespaceEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *FunctionNamespaceEnumerator) SupportedType() resource.ResourceType {
	return scaleway.ScalewayFunctionNamespaceResourceType
}

func (e *FunctionNamespaceEnumerator) Enumerate() ([]*resource.Resource, error) {
	namespaces, err := e.repository.ListAllNamespaces()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(namespaces))

	for _, namespace := range namespaces {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				getRegionalID(namespace.Region.String(), namespace.ID),
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

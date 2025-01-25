package scaleway

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const ScalewayFunctionNamespaceResourceType = "scaleway_function_namespace"

func initScalewayFunctionNamespace(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(ScalewayFunctionNamespaceResourceType, resource.FlagDeepMode)
}

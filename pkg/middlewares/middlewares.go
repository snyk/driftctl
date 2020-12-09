package middlewares

import "github.com/cloudskiff/driftctl/pkg/resource"

type Middleware interface {
	Execute(remoteResources, resourcesFromState *[]resource.Resource) error
}

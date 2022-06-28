package middlewares

import "github.com/snyk/driftctl/enumeration/resource"

type Middleware interface {
	Execute(remoteResources, resourcesFromState *[]*resource.Resource) error
}

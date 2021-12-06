package middlewares

import "github.com/snyk/driftctl/pkg/resource"

type Middleware interface {
	Execute(remoteResources, resourcesFromState *[]*resource.Resource) error
}

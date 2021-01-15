package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

func (r *AwsRouteTable) NormalizeForState() (resource.Resource, error) {
	if r.PropagatingVgws != nil && len(*r.PropagatingVgws) == 0 {
		r.PropagatingVgws = nil
	}
	if r.Route != nil && len(*r.Route) == 0 {
		r.Route = nil
	}
	return r, nil
}

func (r *AwsRouteTable) NormalizeForProvider() (resource.Resource, error) {
	if r.PropagatingVgws != nil && len(*r.PropagatingVgws) == 0 {
		r.PropagatingVgws = nil
	}
	// We do not need route attribute as routes inside tables are expanded to dedicated
	// aws_route resources in resource supplier.
	// When reading state file, the same behavior is applied using a middleware
	r.Route = nil
	return r, nil
}

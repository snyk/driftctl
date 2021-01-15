package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

func (r *AwsDefaultRouteTable) NormalizeForState() (resource.Resource, error) {
	if r.PropagatingVgws != nil && len(*r.PropagatingVgws) == 0 {
		r.PropagatingVgws = nil
	}
	if r.Route != nil && len(*r.Route) == 0 {
		r.Route = nil
	}
	return r, nil
}

func (r *AwsDefaultRouteTable) NormalizeForProvider() (resource.Resource, error) {
	if r.PropagatingVgws != nil && len(*r.PropagatingVgws) == 0 {
		r.PropagatingVgws = nil
	}
	r.Route = nil
	return r, nil
}

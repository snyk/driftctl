package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

func (r *AwsSecurityGroup) NormalizeForState() (resource.Resource, error) {
	//TODO We need to find a way to warn users that some rules in their states could be unmanaged
	if r.Ingress != nil {
		r.Ingress = nil
	}
	if r.Egress != nil {
		r.Egress = nil
	}
	return r, nil
}

func (r *AwsSecurityGroup) NormalizeForProvider() (resource.Resource, error) {
	if r.Ingress != nil {
		r.Ingress = nil
	}
	if r.Egress != nil {
		r.Egress = nil
	}
	return r, nil
}

package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

func (r *AwsDefaultSecurityGroup) NormalizeForState() (resource.Resource, error) {
	if r.Ingress != nil {
		r.Ingress = nil
	}
	if r.Egress != nil {
		r.Egress = nil
	}
	return r, nil
}

func (r *AwsDefaultSecurityGroup) NormalizeForProvider() (resource.Resource, error) {
	if r.Ingress != nil {
		r.Ingress = nil
	}
	if r.Egress != nil {
		r.Egress = nil
	}
	return r, nil
}

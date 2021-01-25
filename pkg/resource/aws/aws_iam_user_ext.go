package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

func (r *AwsIamUser) NormalizeForState() (resource.Resource, error) {
	if r.PermissionsBoundary != nil && *r.PermissionsBoundary == "" {
		r.PermissionsBoundary = nil
	}
	return r, nil
}

func (r *AwsIamUser) NormalizeForProvider() (resource.Resource, error) {
	return r, nil
}

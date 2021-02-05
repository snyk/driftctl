package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

func (r *AwsSqsQueue) NormalizeForState() (resource.Resource, error) {
	if r.Policy != nil && *r.Policy == "" {
		r.Policy = nil
	}
	return r, nil
}

func (r *AwsSqsQueue) NormalizeForProvider() (resource.Resource, error) {
	r.Policy = nil
	return r, nil
}

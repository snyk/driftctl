package aws

import (
	"github.com/cloudskiff/driftctl/pkg/helpers"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *AwsIamPolicy) NormalizeForState() (resource.Resource, error) {
	err := r.normalizePolicy()
	return r, err
}

func (r *AwsIamPolicy) NormalizeForProvider() (resource.Resource, error) {
	err := r.normalizePolicy()
	return r, err
}

func (r *AwsIamPolicy) normalizePolicy() error {
	if r.Policy != nil {
		jsonString, err := helpers.NormalizeJsonString(*r.Policy)
		if err != nil {
			return err
		}
		r.Policy = &jsonString
	}
	return nil
}

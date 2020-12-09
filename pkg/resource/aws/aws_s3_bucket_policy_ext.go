package aws

import (
	"github.com/cloudskiff/driftctl/pkg/helpers"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r AwsS3BucketPolicy) NormalizeForState() (resource.Resource, error) {
	err := r.normalizePolicy()
	return &r, err
}

func (r AwsS3BucketPolicy) NormalizeForProvider() (resource.Resource, error) {
	err := r.normalizePolicy()
	return &r, err
}

func (r *AwsS3BucketPolicy) normalizePolicy() error {
	if r.Policy != nil {
		jsonString, err := helpers.NormalizeJsonString(*r.Policy)
		if err != nil {
			return err
		}
		r.Policy = &jsonString
	}
	return nil
}

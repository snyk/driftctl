package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *AwsIamPolicyAttachment) NormalizeForState() (resource.Resource, error) {
	return r, nil
}

func (r *AwsIamPolicyAttachment) NormalizeForProvider() (resource.Resource, error) {
	if r.Groups != nil && len(*r.Groups) == 0 {
		r.Groups = nil
	}
	if r.Users != nil && len(*r.Users) == 0 {
		r.Users = nil
	}
	return r, nil
}

package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

func (r *AwsIamAccessKey) NormalizeForState() (resource.Resource, error) {
	// As we can't read secrets from aws API once access_key created we need to set
	// fields retrieved from state to nil to avoid drift
	// We can't detect drift if we cannot retrieve latest value from aws API for fields like secrets, passwords etc ...
	r.Secret = nil
	r.SesSmtpPasswordV4 = nil
	return r, nil
}

func (r *AwsIamAccessKey) NormalizeForProvider() (resource.Resource, error) {
	return r, nil
}

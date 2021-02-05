package aws

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *AwsSnsTopic) NormalizeForState() (resource.Resource, error) {
	if r.Policy != nil && *r.Policy == "" {
		r.Policy = nil
	}
	return r, nil
}

func (r *AwsSnsTopic) NormalizeForProvider() (resource.Resource, error) {
	r.Policy = nil
	return r, nil
}

func (r *AwsSnsTopic) String() string {
	if r.DisplayName != nil && *r.DisplayName != "" && r.Name != nil && *r.Name != "" {
		return fmt.Sprintf("%s (%s)", *r.DisplayName, *r.Name)
	}
	if r.Name != nil && *r.Name != "" {
		return *r.Name
	}
	return r.Id
}

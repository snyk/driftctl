package aws

import (
	"fmt"
)

func (r *AwsSnsTopic) String() string {
	if r.DisplayName != nil && *r.DisplayName != "" && r.Name != nil && *r.Name != "" {
		return fmt.Sprintf("%s (%s)", *r.DisplayName, *r.Name)
	}
	if r.Name != nil && *r.Name != "" {
		return *r.Name
	}
	return r.Id
}

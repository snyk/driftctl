package aws

import (
	"fmt"
)

func (r *AwsRoute53Zone) String() string {
	if r.Name == nil {
		return r.TerraformId()
	}
	return fmt.Sprintf("%s (Id: %s)", *r.Name, r.TerraformId())
}

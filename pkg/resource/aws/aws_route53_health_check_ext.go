package aws

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *AwsRoute53HealthCheck) String() string {
	str := r.Id

	name, hasName := r.Tags["Name"]
	if hasName {
		str = name
	}

	if r.Fqdn != nil && *r.Fqdn != "" {
		str += fmt.Sprintf(" (fqdn: %s", *r.Fqdn)
		str = r.addPortAndResPathString(str)
		str += ")"
	}

	if r.IpAddress != nil && *r.IpAddress != "" {
		str += fmt.Sprintf(" (ip: %s", *r.IpAddress)
		str = r.addPortAndResPathString(str)
		str += ")"
	}

	return str
}

func (r *AwsRoute53HealthCheck) addPortAndResPathString(str string) string {
	if r.Port != nil {
		str += fmt.Sprintf(", port: %d", *r.Port)
	}
	if r.ResourcePath != nil {
		str += fmt.Sprintf(", path: %s", *r.ResourcePath)
	}
	return str
}

func (r *AwsRoute53HealthCheck) NormalizeForState() (resource.Resource, error) {
	r.ChildHealthchecks = &[]string{}
	r.Regions = &[]string{}
	return r, nil
}

func (r *AwsRoute53HealthCheck) NormalizeForProvider() (resource.Resource, error) {
	return r, nil
}

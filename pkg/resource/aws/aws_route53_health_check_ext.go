package aws

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *AwsRoute53HealthCheck) Attributes() map[string]string {
	attrs := make(map[string]string)
	name, hasName := r.Tags["Name"]
	if hasName {
		attrs["Name"] = name
	}
	if r.Fqdn != nil && *r.Fqdn != "" {
		attrs["Fqdn"] = *r.Fqdn
		r.addPortAndResPathString(attrs)
	}
	if r.IpAddress != nil && *r.IpAddress != "" {
		attrs["IpAddress"] = *r.IpAddress
		r.addPortAndResPathString(attrs)
	}
	return attrs
}

func (r *AwsRoute53HealthCheck) addPortAndResPathString(attrs map[string]string) {
	if r.Port != nil {
		attrs["Port"] = fmt.Sprintf("%d", *r.Port)
	}
	if r.ResourcePath != nil {
		attrs["Path"] = *r.ResourcePath
	}
}

func (r *AwsRoute53HealthCheck) NormalizeForState() (resource.Resource, error) {
	r.ChildHealthchecks = &[]string{}
	r.Regions = &[]string{}
	return r, nil
}

func (r *AwsRoute53HealthCheck) NormalizeForProvider() (resource.Resource, error) {
	return r, nil
}

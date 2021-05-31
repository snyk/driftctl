package aws

func (r *AwsRoute53Zone) Attributes() map[string]string {
	attrs := make(map[string]string)
	if r.Name != nil && *r.Name != "" {
		attrs["Name"] = *r.Name
	}
	return attrs
}

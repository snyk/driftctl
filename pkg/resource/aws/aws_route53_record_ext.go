package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *AwsRoute53Record) NormalizeForState() (resource.Resource, error) {

	// We read empty array from state but we got nil from cloud provider reads
	if r.Alias != nil && len(*r.Alias) == 0 {
		r.Alias = nil
	}

	// On first run, this field is set to null in state file and to "" after one refresh or apply
	// This ensures that if we find a nil value we don't drift
	if r.HealthCheckId == nil {
		r.HealthCheckId = aws.String("")
	}

	// On first run, this field is set to null in state file and to "" after one refresh or apply
	// This ensures that if we find a nil value we don't drift
	if r.SetIdentifier == nil {
		r.SetIdentifier = aws.String("")
	}

	// This ensures that if we find an empty records value we don't drift
	if r.Records != nil && len(*r.Records) == 0 {
		r.Records = nil
	}

	// This ensures that if we find a nil value we don't drift
	if r.Ttl == nil {
		r.Ttl = aws.Int(0)
	}

	// Since AWS returns the FQDN as the name of the remote record, we must change the Id of the
	// state record to be equivalent (ZoneId_FQDN_Type_SetIdentifier)
	// For a TXT record toto for zone example.com with Id 1234
	// From AWS provider, we retrieve: 1234_toto.example.com_TXT
	// From Terraform state, we retrieve: 1234_toto_TXT
	vars := []string{
		*r.ZoneId,
		*r.Fqdn,
		*r.Type,
	}
	if r.SetIdentifier != nil && *r.SetIdentifier != "" {
		vars = append(vars, *r.SetIdentifier)
	}
	r.Id = strings.Join(vars, "_")

	return r, nil
}

func (r *AwsRoute53Record) NormalizeForProvider() (resource.Resource, error) {

	// This ensures that if we find a nil value we don't drift
	if r.Records != nil && len(*r.Records) == 0 {
		r.Records = nil
	}

	return r, nil
}

func (r *AwsRoute53Record) Attributes() map[string]string {
	attrs := make(map[string]string)
	if r.Fqdn != nil && *r.Fqdn != "" {
		attrs["Fqdn"] = *r.Fqdn
	}
	if r.Type != nil && *r.Type != "" {
		attrs["Type"] = *r.Type
	}
	if r.ZoneId != nil && *r.ZoneId != "" {
		attrs["ZoneId"] = *r.ZoneId
	}
	return attrs
}

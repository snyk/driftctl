package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *AwsRoute53Record) NormalizeForState() (resource.Resource, error) {

	// We read empty array from state but we got nil from cloud provider reads
	if r.Alias != nil && len(*r.Alias) == 0 {
		r.Alias = nil
	}

	// On first run, this field is set to null in state file and to "" after one refresh or apply
	// This ensure that if we find a nil value we dont drift
	if r.HealthCheckId == nil {
		r.HealthCheckId = aws.String("")
	}

	// On first run, this field is set to null in state file and to "" after one refresh or apply
	// This ensure that if we find a nil value we dont drift
	if r.SetIdentifier == nil {
		r.SetIdentifier = aws.String("")
	}

	return r, nil
}

func (r *AwsRoute53Record) NormalizeForProvider() (resource.Resource, error) {
	return r, nil
}

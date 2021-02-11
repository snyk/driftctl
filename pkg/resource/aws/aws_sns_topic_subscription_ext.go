package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/helpers"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *AwsSnsTopicSubscription) NormalizeForState() (resource.Resource, error) {
	err := r.normalizePolicies()
	return r, err
}

func (r *AwsSnsTopicSubscription) NormalizeForProvider() (resource.Resource, error) {
	err := r.normalizePolicies()

	if r.ConfirmationTimeoutInMinutes == nil {
		r.ConfirmationTimeoutInMinutes = aws.Int(1)
	}

	if r.EndpointAutoConfirms == nil {
		r.EndpointAutoConfirms = aws.Bool(false)
	}

	return r, err
}

func (r *AwsSnsTopicSubscription) normalizePolicies() error {
	if r.FilterPolicy != nil {
		jsonString, err := helpers.NormalizeJsonString(*r.FilterPolicy)
		if err != nil {
			return err
		}
		r.FilterPolicy = &jsonString
	}
	if r.DeliveryPolicy != nil {
		jsonString, err := helpers.NormalizeJsonString(*r.DeliveryPolicy)
		if err != nil {
			return err
		}
		r.DeliveryPolicy = &jsonString
	}
	return nil
}

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/resource"
)

func (r *AwsSnsTopic) NormalizeForState() (resource.Resource, error) {
	if r.Policy != nil && *r.Policy == "" {
		r.Policy = nil
	}
	r.normalizeDefaultFields()
	return r, nil
}

func (r *AwsSnsTopic) NormalizeForProvider() (resource.Resource, error) {
	r.Policy = nil
	r.normalizeDefaultFields()
	return r, nil
}

func (r *AwsSnsTopic) normalizeDefaultFields() {
	if r.SqsSuccessFeedbackSampleRate == nil {
		r.SqsSuccessFeedbackSampleRate = aws.Int(0)
	}
	if r.LambdaSuccessFeedbackSampleRate == nil {
		r.LambdaSuccessFeedbackSampleRate = aws.Int(0)
	}
	if r.HttpSuccessFeedbackSampleRate == nil {
		r.HttpSuccessFeedbackSampleRate = aws.Int(0)
	}
	if r.ApplicationSuccessFeedbackSampleRate == nil {
		r.ApplicationSuccessFeedbackSampleRate = aws.Int(0)
	}
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

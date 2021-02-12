package remote

import (
	"errors"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws/awserr"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/alerter"
)

func TestHandleListAwsError(t *testing.T) {

	tests := []struct {
		name       string
		err        error
		wantAlerts alerter.Alerts
		wantErr    bool
	}{
		{
			name:       "Handled error 403",
			err:        remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 403, ""), resourceaws.AwsVpcResourceType),
			wantAlerts: alerter.Alerts{"aws_vpc": []alerter.Alert{NewEnumerationAccessDeniedAlert("aws_vpc", "aws_vpc")}},
			wantErr:    false,
		},
		{
			name:       "Handled error AccessDenied",
			err:        remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, ""), resourceaws.AwsDynamodbTableResourceType),
			wantAlerts: alerter.Alerts{"aws_dynamodb_table": []alerter.Alert{NewEnumerationAccessDeniedAlert("aws_dynamodb_table", "aws_dynamodb_table")}},
			wantErr:    false,
		},
		{
			name:       "Not Handled error code",
			err:        remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 404, ""), resourceaws.AwsVpcResourceType),
			wantAlerts: map[string][]alerter.Alert{},
			wantErr:    true,
		},
		{
			name:       "Not Handled supplier error",
			err:        remoteerror.NewSupplierError(awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 403, ""), map[string]string{}, resourceaws.AwsVpcResourceType),
			wantAlerts: map[string][]alerter.Alert{},
			wantErr:    true,
		},
		{
			name:       "Not Handled error type",
			err:        errors.New("error"),
			wantAlerts: map[string][]alerter.Alert{},
			wantErr:    true,
		},
		{
			name:       "Not Handled root error type",
			err:        remoteerror.NewResourceEnumerationError(errors.New("error"), resourceaws.AwsVpcResourceType),
			wantAlerts: map[string][]alerter.Alert{},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alertr := alerter.NewAlerter()
			gotErr := HandleResourceEnumerationError(tt.err, alertr)
			assert.Equal(t, tt.wantErr, gotErr != nil)

			retrieve := alertr.Retrieve()
			assert.Equal(t, tt.wantAlerts, retrieve)

		})
	}
}

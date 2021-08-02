package remote

import (
	"errors"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/remote/github"
	resourcegithub "github.com/cloudskiff/driftctl/pkg/resource/github"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws/awserr"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/alerter"
)

func TestHandleAwsEnumerationErrors(t *testing.T) {

	tests := []struct {
		name       string
		err        error
		wantAlerts alerter.Alerts
		wantErr    bool
	}{
		{
			name:       "Handled error 403",
			err:        remoteerror.NewResourceScanningError(awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 403, ""), resourceaws.AwsVpcResourceType),
			wantAlerts: alerter.Alerts{"aws_vpc": []alerter.Alert{NewRemoteAccessDeniedAlert(aws.RemoteAWSTerraform, "aws_vpc", "aws_vpc", EnumerationPhase)}},
			wantErr:    false,
		},
		{
			name:       "Handled error AccessDenied",
			err:        remoteerror.NewResourceScanningError(awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, ""), resourceaws.AwsDynamodbTableResourceType),
			wantAlerts: alerter.Alerts{"aws_dynamodb_table": []alerter.Alert{NewRemoteAccessDeniedAlert(aws.RemoteAWSTerraform, "aws_dynamodb_table", "aws_dynamodb_table", EnumerationPhase)}},
			wantErr:    false,
		},
		{
			name:       "Not Handled error code",
			err:        remoteerror.NewResourceScanningError(awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 404, ""), resourceaws.AwsVpcResourceType),
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
			err:        remoteerror.NewResourceScanningError(errors.New("error"), resourceaws.AwsVpcResourceType),
			wantAlerts: map[string][]alerter.Alert{},
			wantErr:    true,
		},
		{
			name:       "Handle AccessDenied error",
			err:        remoteerror.NewResourceScanningError(errors.New("an error occured: AccessDenied: 403"), resourceaws.AwsVpcResourceType),
			wantAlerts: alerter.Alerts{"aws_vpc": []alerter.Alert{NewRemoteAccessDeniedAlert(aws.RemoteAWSTerraform, "aws_vpc", "aws_vpc", EnumerationPhase)}},
			wantErr:    false,
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

func TestHandleGithubEnumerationErrors(t *testing.T) {

	tests := []struct {
		name       string
		err        error
		wantAlerts alerter.Alerts
		wantErr    bool
	}{
		{
			name:       "Handled graphql error",
			err:        remoteerror.NewResourceScanningError(errors.New("Your token has not been granted the required scopes to execute this query."), resourcegithub.GithubTeamResourceType),
			wantAlerts: alerter.Alerts{"github_team": []alerter.Alert{NewRemoteAccessDeniedAlert(github.RemoteGithubTerraform, "github_team", "github_team", EnumerationPhase)}},
			wantErr:    false,
		},
		{
			name:       "Not handled graphql error",
			err:        remoteerror.NewResourceScanningError(errors.New("This is a not handler graphql error"), resourcegithub.GithubTeamResourceType),
			wantAlerts: map[string][]alerter.Alert{},
			wantErr:    true,
		},
		{
			name:       "Not Handled supplier error",
			err:        remoteerror.NewSupplierError(errors.New("An error from the supplier"), map[string]string{}, resourcegithub.GithubTeamResourceType),
			wantAlerts: map[string][]alerter.Alert{},
			wantErr:    true,
		},
		{
			name:       "Not Handled error type",
			err:        errors.New("error"),
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

func TestHandleAwsDetailsFetchingErrors(t *testing.T) {

	tests := []struct {
		name       string
		err        error
		wantAlerts alerter.Alerts
		wantErr    bool
	}{
		{
			name:       "Handle AccessDeniedException error",
			err:        remoteerror.NewResourceScanningError(awserr.NewRequestFailure(awserr.New("AccessDeniedException", "test", errors.New("")), 403, ""), resourceaws.AwsVpcResourceType),
			wantAlerts: alerter.Alerts{"aws_vpc": []alerter.Alert{NewRemoteAccessDeniedAlert(aws.RemoteAWSTerraform, "aws_vpc", "aws_vpc", DetailsFetchingPhase)}},
			wantErr:    false,
		},
		{
			name:       "Handle AccessDenied error",
			err:        remoteerror.NewResourceScanningError(awserr.NewRequestFailure(awserr.New("test", "error: AccessDenied", errors.New("")), 403, ""), resourceaws.AwsVpcResourceType),
			wantAlerts: alerter.Alerts{"aws_vpc": []alerter.Alert{NewRemoteAccessDeniedAlert(aws.RemoteAWSTerraform, "aws_vpc", "aws_vpc", DetailsFetchingPhase)}},
			wantErr:    false,
		},
		{
			name:       "Handle AuthorizationError error",
			err:        remoteerror.NewResourceScanningError(awserr.NewRequestFailure(awserr.New("test", "error: AuthorizationError", errors.New("")), 403, ""), resourceaws.AwsVpcResourceType),
			wantAlerts: alerter.Alerts{"aws_vpc": []alerter.Alert{NewRemoteAccessDeniedAlert(aws.RemoteAWSTerraform, "aws_vpc", "aws_vpc", DetailsFetchingPhase)}},
			wantErr:    false,
		},
		{
			name:       "Unhandled error",
			err:        remoteerror.NewResourceScanningError(awserr.NewRequestFailure(awserr.New("test", "error: dummy error", errors.New("")), 403, ""), resourceaws.AwsVpcResourceType),
			wantAlerts: alerter.Alerts{},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alertr := alerter.NewAlerter()
			gotErr := HandleResourceDetailsFetchingError(tt.err, alertr)
			assert.Equal(t, tt.wantErr, gotErr != nil)

			retrieve := alertr.Retrieve()
			assert.Equal(t, tt.wantAlerts, retrieve)

		})
	}
}

func TestEnumerationAccessDeniedAlert_GetProviderMessage(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     string
	}{
		{
			name:     "test for unsupported provider",
			provider: "foobar",
			want:     "",
		},
		{
			name:     "test for AWS",
			provider: aws.RemoteAWSTerraform,
			want:     "It seems that we got access denied exceptions while listing resources.\nThe latest minimal read-only IAM policy for driftctl is always available here, please update yours: https://docs.driftctl.com/aws/policy",
		},
		{
			name:     "test for github",
			provider: github.RemoteGithubTerraform,
			want:     "It seems that we got access denied exceptions while listing resources.\nPlease be sure that your Github token has the right permissions, check the last up-to-date documentation there: https://docs.driftctl.com/github/policy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewRemoteAccessDeniedAlert(tt.provider, "supplier_type", "listed_type_error", EnumerationPhase)
			if got := e.GetProviderMessage(); got != tt.want {
				t.Errorf("GetProviderMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetailsFetchingAccessDeniedAlert_GetProviderMessage(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     string
	}{
		{
			name:     "test for unsupported provider",
			provider: "foobar",
			want:     "",
		},
		{
			name:     "test for AWS",
			provider: aws.RemoteAWSTerraform,
			want:     "It seems that we got access denied exceptions while reading details of resources.\nThe latest minimal read-only IAM policy for driftctl is always available here, please update yours: https://docs.driftctl.com/aws/policy",
		},
		{
			name:     "test for github",
			provider: github.RemoteGithubTerraform,
			want:     "It seems that we got access denied exceptions while reading details of resources.\nPlease be sure that your Github token has the right permissions, check the last up-to-date documentation there: https://docs.driftctl.com/github/policy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewRemoteAccessDeniedAlert(tt.provider, "supplier_type", "listed_type_error", DetailsFetchingPhase)
			if got := e.GetProviderMessage(); got != tt.want {
				t.Errorf("GetProviderMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

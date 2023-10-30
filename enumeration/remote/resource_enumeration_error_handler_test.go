package remote

import (
	"errors"
	"testing"

	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"

	resourcegithub "github.com/snyk/driftctl/enumeration/resource/github"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws/awserr"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
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
			err:        remoteerr.NewResourceListingError(awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 403, ""), resourceaws.AwsVpcResourceType),
			wantAlerts: alerter.Alerts{"aws_vpc": []alerter.Alert{alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 403, ""), "aws_vpc", "aws_vpc"), alerts.EnumerationPhase)}},
			wantErr:    false,
		},
		{
			name:       "Handled error AccessDenied",
			err:        remoteerr.NewResourceListingError(awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, ""), resourceaws.AwsDynamodbTableResourceType),
			wantAlerts: alerter.Alerts{"aws_dynamodb_table": []alerter.Alert{alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, ""), "aws_dynamodb_table", "aws_dynamodb_table"), alerts.EnumerationPhase)}},
			wantErr:    false,
		},
		{
			name:       "Not Handled error code",
			err:        remoteerr.NewResourceListingError(awserr.NewRequestFailure(awserr.New("", "", errors.New("")), 404, ""), resourceaws.AwsVpcResourceType),
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
			err:        remoteerr.NewResourceListingError(errors.New("error"), resourceaws.AwsVpcResourceType),
			wantAlerts: map[string][]alerter.Alert{},
			wantErr:    true,
		},
		{
			name:       "Handle AccessDenied error",
			err:        remoteerr.NewResourceListingError(errors.New("an error occured: AccessDenied: 403"), resourceaws.AwsVpcResourceType),
			wantAlerts: alerter.Alerts{"aws_vpc": []alerter.Alert{alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("an error occured: AccessDenied: 403"), "aws_vpc", "aws_vpc"), alerts.EnumerationPhase)}},
			wantErr:    false,
		},
		{
			name:       "Access denied error on a single resource",
			err:        remoteerr.NewResourceScanningError(errors.New("Error: AccessDenied: 403 ..."), resourceaws.AwsS3BucketResourceType, "my-bucket"),
			wantAlerts: alerter.Alerts{"aws_s3_bucket.my-bucket": []alerter.Alert{alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("Error: AccessDenied: 403 ..."), "aws_s3_bucket.my-bucket", "aws_s3_bucket"), alerts.EnumerationPhase)}},
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
			err:        remoteerr.NewResourceListingError(errors.New("Your token has not been granted the required scopes to execute this query."), resourcegithub.GithubTeamResourceType),
			wantAlerts: alerter.Alerts{"github_team": []alerter.Alert{alerts.NewRemoteAccessDeniedAlert(common.RemoteGithubTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("Your token has not been granted the required scopes to execute this query."), "github_team", "github_team"), alerts.EnumerationPhase)}},
			wantErr:    false,
		},
		{
			name:       "Not handled graphql error",
			err:        remoteerr.NewResourceListingError(errors.New("This is a not handler graphql error"), resourcegithub.GithubTeamResourceType),
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

func TestHandleGoogleEnumerationErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantAlerts alerter.Alerts
		wantErr    bool
	}{
		{
			name:       "Handled 403 error",
			err:        remoteerr.NewResourceListingError(status.Error(codes.PermissionDenied, "useless message"), "google_type"),
			wantAlerts: alerter.Alerts{"google_type": []alerter.Alert{alerts.NewRemoteAccessDeniedAlert(common.RemoteGoogleTerraform, remoteerr.NewResourceListingErrorWithType(status.Error(codes.PermissionDenied, "useless message"), "google_type", "google_type"), alerts.EnumerationPhase)}},
			wantErr:    false,
		},
		{
			name:       "Not handled non 403 error",
			err:        status.Error(codes.Unknown, ""),
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
			provider: common.RemoteAWSTerraform,
			want:     "It seems that we got access denied exceptions while listing resources.\nThe latest minimal read-only IAM policy for driftctl is always available here, please update yours: https://docs.driftctl.com/aws/policy",
		},
		{
			name:     "test for github",
			provider: common.RemoteGithubTerraform,
			want:     "It seems that we got access denied exceptions while listing resources.\nPlease be sure that your Github token has the right permissions, check the last up-to-date documentation there: https://docs.driftctl.com/github/policy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := alerts.NewRemoteAccessDeniedAlert(tt.provider, remoteerr.NewResourceListingErrorWithType(errors.New("dummy error"), "supplier_type", "listed_type_error"), alerts.EnumerationPhase)
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
			provider: common.RemoteAWSTerraform,
			want:     "It seems that we got access denied exceptions while reading details of resources.\nThe latest minimal read-only IAM policy for driftctl is always available here, please update yours: https://docs.driftctl.com/aws/policy",
		},
		{
			name:     "test for github",
			provider: common.RemoteGithubTerraform,
			want:     "It seems that we got access denied exceptions while reading details of resources.\nPlease be sure that your Github token has the right permissions, check the last up-to-date documentation there: https://docs.driftctl.com/github/policy",
		},
		{
			name:     "test for google",
			provider: common.RemoteGoogleTerraform,
			want:     "It seems that we got access denied exceptions while reading details of resources.\nPlease ensure that you have configured the required roles, please check our documentation at https://docs.driftctl.com/google/policy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := alerts.NewRemoteAccessDeniedAlert(tt.provider, remoteerr.NewResourceListingErrorWithType(errors.New("dummy error"), "supplier_type", "listed_type_error"), alerts.DetailsFetchingPhase)
			if got := e.GetProviderMessage(); got != tt.want {
				t.Errorf("GetProviderMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceScanningErrorMethods(t *testing.T) {

	tests := []struct {
		name                 string
		err                  *remoteerr.ResourceScanningError
		expectedError        string
		expectedResourceType string
	}{
		{
			name:                 "Handled error AccessDenied",
			err:                  remoteerr.NewResourceListingError(awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, ""), resourceaws.AwsDynamodbTableResourceType),
			expectedError:        "error scanning resource type aws_dynamodb_table: AccessDeniedException: \n\tstatus code: 403, request id: \ncaused by: ",
			expectedResourceType: resourceaws.AwsDynamodbTableResourceType,
		},
		{
			name:                 "Handle AccessDenied error",
			err:                  remoteerr.NewResourceListingError(errors.New("an error occured: AccessDenied: 403"), resourceaws.AwsVpcResourceType),
			expectedError:        "error scanning resource type aws_vpc: an error occured: AccessDenied: 403",
			expectedResourceType: resourceaws.AwsVpcResourceType,
		},
		{
			name:                 "Access denied error on a single resource",
			err:                  remoteerr.NewResourceScanningError(errors.New("Error: AccessDenied: 403 ..."), resourceaws.AwsS3BucketResourceType, "my-bucket"),
			expectedError:        "error scanning resource aws_s3_bucket.my-bucket: Error: AccessDenied: 403 ...",
			expectedResourceType: resourceaws.AwsS3BucketResourceType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedError, tt.err.Error())
			assert.Equal(t, tt.expectedResourceType, tt.err.ResourceType())
		})
	}
}

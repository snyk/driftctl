package remote

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/remote/alerts"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
)

func HandleResourceEnumerationError(err error, alerter alerter.AlerterInterface) error {
	listError, ok := err.(*remoteerror.ResourceScanningError)
	if !ok {
		return err
	}

	rootCause := listError.RootCause()

	reqerr, ok := rootCause.(awserr.RequestFailure)
	if ok {
		return handleAWSError(alerter, listError, reqerr)
	}

	// This handles access denied errors like the following:
	// aws_s3_bucket_policy: AccessDenied: Error listing bucket policy <policy_name>
	if strings.Contains(rootCause.Error(), "AccessDenied") {
		alerts.SendEnumerationAlert(common.RemoteAWSTerraform, alerter, listError)
		return nil
	}

	if strings.HasPrefix(
		rootCause.Error(),
		"Your token has not been granted the required scopes to execute this query.",
	) {
		alerts.SendEnumerationAlert(common.RemoteGithubTerraform, alerter, listError)
		return nil
	}

	return err
}

func HandleResourceDetailsFetchingError(err error, alerter alerter.AlerterInterface) error {
	listError, ok := err.(*remoteerror.ResourceScanningError)
	if !ok {
		return err
	}

	rootCause := listError.RootCause()

	// This handles access denied errors like the following:
	// iam_role_policy: error reading IAM Role Policy (<policy>): AccessDenied: User: <role_arn> ...
	if strings.HasPrefix(rootCause.Error(), "AccessDeniedException") ||
		strings.Contains(rootCause.Error(), "AccessDenied") ||
		strings.Contains(rootCause.Error(), "AuthorizationError") {
		alerts.SendDetailsFetchingAlert(common.RemoteAWSTerraform, alerter, listError)
		return nil
	}

	return err
}

func handleAWSError(alerter alerter.AlerterInterface, listError *remoteerror.ResourceScanningError, reqerr awserr.RequestFailure) error {
	if reqerr.StatusCode() == 403 || (reqerr.StatusCode() == 400 && strings.Contains(reqerr.Code(), "AccessDenied")) {
		alerts.SendEnumerationAlert(common.RemoteAWSTerraform, alerter, listError)
		return nil
	}

	return reqerr
}

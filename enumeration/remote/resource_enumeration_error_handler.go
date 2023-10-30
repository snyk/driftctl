package remote

import (
	"strings"

	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleResourceEnumerationError(err error, alerter alerter.AlerterInterface) error {
	listError, ok := err.(*remoteerror.ResourceScanningError)
	if !ok {
		return err
	}

	rootCause := listError.RootCause()

	// We cannot use the status.FromError() method because AWS errors are not well-formed.
	// Indeed, they compose the error interface without implementing the Error() method and thus triggering a nil panic
	// when returning an unknown error from status.FromError()
	// As a workaround we duplicated the logic from status.FromError here
	if _, ok := rootCause.(interface{ GRPCStatus() *status.Status }); ok {
		return handleGoogleEnumerationError(alerter, listError, status.Convert(rootCause))
	}

	// at least for storage api google sdk does not return grpc error so we parse the error message.
	if shouldHandleGoogleForbiddenError(listError) {
		alerts.SendEnumerationAlert(common.RemoteGoogleTerraform, alerter, listError)
		return nil
	}

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

func handleAWSError(alerter alerter.AlerterInterface, listError *remoteerror.ResourceScanningError, reqerr awserr.RequestFailure) error {
	if reqerr.StatusCode() == 403 || (reqerr.StatusCode() == 400 && strings.Contains(reqerr.Code(), "AccessDenied")) {
		alerts.SendEnumerationAlert(common.RemoteAWSTerraform, alerter, listError)
		return nil
	}

	return reqerr
}

func handleGoogleEnumerationError(alerter alerter.AlerterInterface, err *remoteerror.ResourceScanningError, st *status.Status) error {
	if st.Code() == codes.PermissionDenied {
		alerts.SendEnumerationAlert(common.RemoteGoogleTerraform, alerter, err)
		return nil
	}
	return err
}

func shouldHandleGoogleForbiddenError(err *remoteerror.ResourceScanningError) bool {
	errMsg := err.RootCause().Error()

	// Check if this is a Google related error
	if !strings.Contains(errMsg, "googleapi") {
		return false
	}

	if strings.Contains(errMsg, "Error 403") {
		return true
	}

	return false
}

package remote

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/remote/github"
	"github.com/sirupsen/logrus"
)

type ScanningPhase int

const (
	EnumerationPhase ScanningPhase = iota
	DetailsFetchingPhase
)

type RemoteAccessDeniedAlert struct {
	message       string
	provider      string
	scanningPhase ScanningPhase
}

func NewRemoteAccessDeniedAlert(provider, supplierType, listedTypeError string, scanningPhase ScanningPhase) *RemoteAccessDeniedAlert {
	var message string
	switch scanningPhase {
	case EnumerationPhase:
		message = fmt.Sprintf("Ignoring %s from drift calculation: Listing %s is forbidden.", supplierType, listedTypeError)
	case DetailsFetchingPhase:
		message = fmt.Sprintf("Ignoring %s from drift calculation: Reading details of %s is forbidden.", supplierType, listedTypeError)
	default:
		message = fmt.Sprintf("Ignoring %s from drift calculation: %s", supplierType, listedTypeError)
	}
	return &RemoteAccessDeniedAlert{message, provider, scanningPhase}
}

func (e *RemoteAccessDeniedAlert) Message() string {
	return e.message
}

func (e *RemoteAccessDeniedAlert) ShouldIgnoreResource() bool {
	return true
}

func (e *RemoteAccessDeniedAlert) GetProviderMessage() string {
	var message string
	if e.scanningPhase == DetailsFetchingPhase {
		message = "It seems that we got access denied exceptions while reading details of resources.\n"
	}
	if e.scanningPhase == EnumerationPhase {
		message = "It seems that we got access denied exceptions while listing resources.\n"
	}

	switch e.provider {
	case github.RemoteGithubTerraform:
		message += "Please be sure that your Github token has the right permissions, check the last up-to-date documentation there: https://docs.driftctl.com/github/policy"
	case aws.RemoteAWSTerraform:
		message += "The latest minimal read-only IAM policy for driftctl is always available here, please update yours: https://docs.driftctl.com/aws/policy"
	default:
		return ""
	}
	return message
}

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
		sendEnumerationAlert(aws.RemoteAWSTerraform, alerter, listError)
		return nil
	}

	if strings.HasPrefix(
		rootCause.Error(),
		"Your token has not been granted the required scopes to execute this query.",
	) {
		sendEnumerationAlert(github.RemoteGithubTerraform, alerter, listError)
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
		sendDetailsFetchingAlert(aws.RemoteAWSTerraform, alerter, listError)
		return nil
	}

	return err
}

func handleAWSError(alerter alerter.AlerterInterface, listError *remoteerror.ResourceScanningError, reqerr awserr.RequestFailure) error {
	if reqerr.StatusCode() == 403 || (reqerr.StatusCode() == 400 && strings.Contains(reqerr.Code(), "AccessDenied")) {
		sendEnumerationAlert(aws.RemoteAWSTerraform, alerter, listError)
		return nil
	}

	return reqerr
}

func sendRemoteAccessDeniedAlert(provider string, alerter alerter.AlerterInterface, listError *remoteerror.ResourceScanningError, p ScanningPhase) {
	logrus.WithFields(logrus.Fields{
		"supplier_type": listError.SupplierType(),
		"listed_type":   listError.ListedTypeError(),
	}).Debugf("Got an access denied error")
	alerter.SendAlert(listError.SupplierType(), NewRemoteAccessDeniedAlert(provider, listError.SupplierType(), listError.ListedTypeError(), p))
}

func sendEnumerationAlert(provider string, alerter alerter.AlerterInterface, listError *remoteerror.ResourceScanningError) {
	sendRemoteAccessDeniedAlert(provider, alerter, listError, EnumerationPhase)
}

func sendDetailsFetchingAlert(provider string, alerter alerter.AlerterInterface, listError *remoteerror.ResourceScanningError) {
	sendRemoteAccessDeniedAlert(provider, alerter, listError, DetailsFetchingPhase)
}

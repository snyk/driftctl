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

type EnumerationAccessDeniedAlert struct {
	message  string
	provider string
}

func NewEnumerationAccessDeniedAlert(provider, supplierType, listedTypeError string) *EnumerationAccessDeniedAlert {
	message := fmt.Sprintf("Ignoring %s from drift calculation: Listing %s is forbidden.", supplierType, listedTypeError)
	return &EnumerationAccessDeniedAlert{message, provider}
}

func (e *EnumerationAccessDeniedAlert) Message() string {
	return e.message
}

func (e *EnumerationAccessDeniedAlert) ShouldIgnoreResource() bool {
	return true
}

func (e *EnumerationAccessDeniedAlert) GetProviderMessage() string {
	message := "It seems that we got access denied exceptions while listing resources.\n"
	switch e.provider {
	case github.RemoteGithubTerraform:
		message += "Please be sure that your Github token has the right permissions, check the last up-to-date documentation there: https://docs.driftctl.com/providers/github/authentication#least-privileged-policy"
	case aws.RemoteAWSTerraform:
		message += "The latest minimal read-only IAM policy for driftctl is always available here, please update yours: https://docs.driftctl.com/providers/aws/authentication#least-privileged-policy"
	default:
		return ""
	}
	return message
}

func HandleResourceEnumerationError(err error, alerter *alerter.Alerter) error {
	listError, ok := err.(*remoteerror.ResourceEnumerationError)
	if !ok {
		return err
	}

	rootCause := listError.RootCause()

	reqerr, ok := rootCause.(awserr.RequestFailure)
	if ok {
		return handleAWSError(alerter, listError, reqerr)
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

func handleAWSError(alerter alerter.AlerterInterface, listError *remoteerror.ResourceEnumerationError, reqerr awserr.RequestFailure) error {
	if reqerr.StatusCode() == 403 || (reqerr.StatusCode() == 400 && strings.Contains(reqerr.Code(), "AccessDenied")) {
		sendEnumerationAlert(aws.RemoteAWSTerraform, alerter, listError)
		return nil
	}

	return reqerr
}

func sendEnumerationAlert(provider string, alerter alerter.AlerterInterface, listError *remoteerror.ResourceEnumerationError) {
	logrus.WithFields(logrus.Fields{
		"supplier_type": listError.SupplierType(),
		"listed_type":   listError.ListedTypeError(),
	}).Debugf("Got an access denied error")
	alerter.SendAlert(listError.SupplierType(), NewEnumerationAccessDeniedAlert(provider, listError.SupplierType(), listError.ListedTypeError()))
}

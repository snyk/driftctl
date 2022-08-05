package alerts

import (
	"fmt"
	"strings"

	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/resource"

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
	resource      *resource.Resource
}

func NewRemoteAccessDeniedAlert(provider string, scanErr *remoteerror.ResourceScanningError, scanningPhase ScanningPhase) *RemoteAccessDeniedAlert {
	var message string
	switch scanningPhase {
	case EnumerationPhase:
		message = fmt.Sprintf(
			"Ignoring %s from drift calculation: Listing %s is forbidden: %s",
			scanErr.Resource(),
			scanErr.ListedTypeError(),
			scanErr.RootCause().Error(),
		)
	case DetailsFetchingPhase:
		message = fmt.Sprintf(
			"Ignoring %s from drift calculation: Reading details of %s is forbidden: %s",
			scanErr.Resource(),
			scanErr.ListedTypeError(),
			scanErr.RootCause().Error(),
		)
	default:
		message = fmt.Sprintf(
			"Ignoring %s from drift calculation: %s",
			scanErr.Resource(),
			scanErr.RootCause().Error(),
		)
	}

	var relatedResource *resource.Resource
	resourceFQDNSSplit := strings.SplitN(scanErr.Resource(), ".", 2)
	if len(resourceFQDNSSplit) == 2 {
		relatedResource = &resource.Resource{
			Id:   resourceFQDNSSplit[1],
			Type: resourceFQDNSSplit[0],
		}
	}

	return &RemoteAccessDeniedAlert{message, provider, scanningPhase, relatedResource}
}

func (e *RemoteAccessDeniedAlert) Message() string {
	return e.message
}

func (e *RemoteAccessDeniedAlert) ShouldIgnoreResource() bool {
	return true
}

func (e *RemoteAccessDeniedAlert) Resource() *resource.Resource {
	return e.resource
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
	case common.RemoteGithubTerraform:
		message += "Please be sure that your Github token has the right permissions, check the last up-to-date documentation there: https://docs.driftctl.com/github/policy"
	case common.RemoteAWSTerraform:
		message += "The latest minimal read-only IAM policy for driftctl is always available here, please update yours: https://docs.driftctl.com/aws/policy"
	case common.RemoteGoogleTerraform:
		message += "Please ensure that you have configured the required roles, please check our documentation at https://docs.driftctl.com/google/policy"
	default:
		return ""
	}
	return message
}

func sendRemoteAccessDeniedAlert(provider string, alerter alerter.AlerterInterface, listError *remoteerror.ResourceScanningError, p ScanningPhase) {
	logrus.WithFields(logrus.Fields{
		"resource":    listError.Resource(),
		"listed_type": listError.ListedTypeError(),
	}).Debugf("Got an access denied error: %+v", listError.Error())
	alerter.SendAlert(listError.Resource(), NewRemoteAccessDeniedAlert(provider, listError, p))
}

func SendEnumerationAlert(provider string, alerter alerter.AlerterInterface, listError *remoteerror.ResourceScanningError) {
	sendRemoteAccessDeniedAlert(provider, alerter, listError, EnumerationPhase)
}

func SendDetailsFetchingAlert(provider string, alerter alerter.AlerterInterface, listError *remoteerror.ResourceScanningError) {
	sendRemoteAccessDeniedAlert(provider, alerter, listError, DetailsFetchingPhase)
}

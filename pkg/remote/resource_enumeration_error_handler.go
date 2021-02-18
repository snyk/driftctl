package remote

import (
	"fmt"
	"strings"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/sirupsen/logrus"
)

type EnumerationAccessDeniedAlert struct {
	message string
}

func NewEnumerationAccessDeniedAlert(supplierType, listedTypeError string) *EnumerationAccessDeniedAlert {
	message := fmt.Sprintf("Ignoring %s from drift calculation: Listing %s is forbidden.", supplierType, listedTypeError)
	return &EnumerationAccessDeniedAlert{message}
}

func (e *EnumerationAccessDeniedAlert) Message() string {
	return e.message
}

func (e *EnumerationAccessDeniedAlert) ShouldIgnoreResource() bool {
	return true
}

func HandleResourceEnumerationError(err error, alertr *alerter.Alerter) error {
	listError, ok := err.(*remoteerror.ResourceEnumerationError)
	if !ok {
		return err
	}

	reqerr, ok := listError.RootCause().(awserr.RequestFailure)
	if !ok {
		return err
	}

	if reqerr.StatusCode() == 403 || (reqerr.StatusCode() == 400 && strings.Contains(reqerr.Code(), "AccessDenied")) {
		logrus.WithFields(logrus.Fields{
			"supplier_type": listError.SupplierType(),
			"listed_type":   listError.ListedTypeError(),
		}).Debugf("Got an access denied error")
		alertr.SendAlert(listError.SupplierType(), NewEnumerationAccessDeniedAlert(listError.SupplierType(), listError.ListedTypeError()))
		return nil
	}

	return err
}

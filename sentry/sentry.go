package sentry

import (
	"fmt"
	"reflect"

	gosentry "github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	cmderrors "github.com/snyk/driftctl/pkg/cmd/errors"
	"github.com/snyk/driftctl/pkg/version"
)

var excludedErrorTypes = []error{
	cmderrors.UsageError{},
}

func Initialize() error {
	logrus.Debug("Enabled error reporting")
	return gosentry.Init(gosentry.ClientOptions{
		Dsn:              "https://9f2b735e20bc452387f7fa093f786173@o495597.ingest.sentry.io/5568568",
		Release:          fmt.Sprintf("driftctl@%s", version.Current()),
		AttachStacktrace: true,
	})
}

func shouldCaptureException(err error) bool {
	errType, causeType := reflect.TypeOf(err), reflect.TypeOf(errors.Cause(err))
	for _, exludedError := range excludedErrorTypes {
		switch reflect.TypeOf(exludedError) {
		case errType:
			return false
		case causeType:
			return false
		default:
		}
	}
	logrus.WithFields(logrus.Fields{
		"error_type": errType,
		"cause_type": causeType,
	}).Debug("Sentry captured error")
	return true
}

func CaptureException(err error) {
	if shouldCaptureException(err) {
		gosentry.CaptureException(err)
	}
}

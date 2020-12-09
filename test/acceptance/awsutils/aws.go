package awsutils

import (
	"github.com/aws/aws-sdk-go/aws/session"
)

func Session() *session.Session {
	return session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
}

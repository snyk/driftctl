package middlewares

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/sirupsen/logrus"
)

type Chain []Middleware

func NewChain(middlewares ...Middleware) Chain {
	return middlewares
}

func (c Chain) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	for _, middleware := range c {
		logrus.WithFields(logrus.Fields{
			"middleware": fmt.Sprintf("%T", middleware),
		}).Debug("Starting middleware")
		err := middleware.Execute(remoteResources, resourcesFromState)
		if err != nil {
			return err
		}
	}
	return nil
}

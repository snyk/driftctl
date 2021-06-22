package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/sirupsen/logrus"
)

// AwsTagsAllField middleware ignore tags_all field on all resources to avoid unexpected noise.
type AwsTagsAllField struct{}

func NewAwsTagsAllField() AwsTagsAllField {
	return AwsTagsAllField{}
}

func (m AwsTagsAllField) Execute(remoteResources, _ *[]resource.Resource) error {

	newRemoteResources := make([]resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		if (*remoteResource.Attributes())["tags_all"] != nil {
			(*remoteResource.Attributes())["tags_all"] = nil

			logrus.WithFields(logrus.Fields{
				"id":   remoteResource.TerraformId(),
				"type": remoteResource.TerraformType(),
			}).Debug("Ignoring tags_all field on resource")
		}

		newRemoteResources = append(newRemoteResources, remoteResource)
	}

	*remoteResources = newRemoteResources

	return nil
}

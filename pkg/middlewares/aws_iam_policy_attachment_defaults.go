package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// Default iam policy attachment should not be shown as unmanaged as they are present by default
// This middleware ignores default iam policy attachment from unmanaged resources if they are not managed by IaC
type AwsIamPolicyAttachmentDefaults struct{}

var ignoredIamPolicyAttachmentIds = []string{
	"AWSServiceRoleForSSO-arn:aws:iam::aws:policy/aws-service-role/AWSSSOServiceRolePolicy",
}

func NewAwsIamPolicyAttachmentDefaults() AwsIamPolicyAttachmentDefaults {
	return AwsIamPolicyAttachmentDefaults{}
}

func (m AwsIamPolicyAttachmentDefaults) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	newRemoteResources := make([]resource.Resource, 0)

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than iam policy attachment
		if remoteResource.TerraformType() != aws.AwsIamPolicyAttachmentResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		existInState := false
		for _, stateResource := range *resourcesFromState {
			if resource.IsSameResource(remoteResource, stateResource) {
				existInState = true
				break
			}
		}

		if existInState {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		isIgnored := false
		for _, id := range ignoredIamPolicyAttachmentIds {
			if remoteResource.TerraformId() == id {
				isIgnored = true
			}
		}

		if !isIgnored {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		logrus.WithFields(logrus.Fields{
			"id":   remoteResource.TerraformId(),
			"type": remoteResource.TerraformType(),
		}).Debug("Ignoring default iam policy attachment as it is not managed by IaC")
	}

	*remoteResources = newRemoteResources

	return nil
}

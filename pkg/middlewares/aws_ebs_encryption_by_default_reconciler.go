package middlewares

import (
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsEbsEncryptionByDefaultReconciler is a middleware that create remote equivalent
// "aws_ebs_encryption_by_default" resources from state resources.
// Since we don't have an ID for remote resources of this type.
type AwsEbsEncryptionByDefaultReconciler struct {
	resourceFactory resource.ResourceFactory
}

func NewAwsEbsEncryptionByDefaultReconciler(resourceFactory resource.ResourceFactory) AwsEbsEncryptionByDefaultReconciler {
	return AwsEbsEncryptionByDefaultReconciler{
		resourceFactory: resourceFactory,
	}
}

func (m AwsEbsEncryptionByDefaultReconciler) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newStateResources := make([]*resource.Resource, 0)
	newRemoteResources := make([]*resource.Resource, 0)

	var defaultEbsEncryption *resource.Resource

	for _, res := range *remoteResources {
		// Ignore all resources other than aws_ebs_encryption_by_default
		if res.ResourceType() != aws.AwsEbsEncryptionByDefaultResourceType {
			newRemoteResources = append(newRemoteResources, res)
			continue
		}
		defaultEbsEncryption = res
		break
	}

	for _, res := range *resourcesFromState {
		newStateResources = append(newStateResources, res)

		// Ignore all resources other than aws_ebs_encryption_by_default
		if res.ResourceType() != aws.AwsEbsEncryptionByDefaultResourceType {
			continue
		}

		// Create the same resource in remote but with the remote attributes, so we can compare it with the state resource
		newRemoteResources = append(newRemoteResources, m.resourceFactory.CreateAbstractResource(
			res.ResourceType(),
			res.ResourceId(),
			map[string]interface{}{
				"id":      res.ResourceId(),
				"enabled": *defaultEbsEncryption.Attributes().GetBool("enabled"),
			},
		))
	}

	*resourcesFromState = newStateResources
	*remoteResources = newRemoteResources
	return nil
}

package middlewares

import (
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsEbsEncryptionByDefaultReconciler is a middleware that either creates an 'aws_ebs_encryption_by_default' resource
// based on its equivalent state one just for the purpose of getting the Terraform custom Id, or removes the resource
// from our list of remote resources if it is not managed and is disabled.
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

	var found bool
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

	// We can encounter this case when we don't have permission to get this setting from AWS.
	if defaultEbsEncryption == nil {
		return nil
	}

	for _, res := range *resourcesFromState {
		newStateResources = append(newStateResources, res)

		// Ignore all resources other than aws_ebs_encryption_by_default
		if res.ResourceType() != aws.AwsEbsEncryptionByDefaultResourceType {
			continue
		}

		// Create a new remote resource that will be similar to the state resource but with the 'enabled' attribute of the remote one.
		// The reason why is that the id is a random string created by Terraform that we need to compare two resources.
		newRemoteResources = append(newRemoteResources, m.resourceFactory.CreateAbstractResource(
			res.ResourceType(),
			res.ResourceId(),
			map[string]interface{}{
				"id":      res.ResourceId(),
				"enabled": *defaultEbsEncryption.Attributes().GetBool("enabled"),
			},
		))
		found = true
	}

	if defaultEbsEncryption != nil && !found && *defaultEbsEncryption.Attributes().GetBool("enabled") {
		newRemoteResources = append(newRemoteResources, defaultEbsEncryption)
	}

	*resourcesFromState = newStateResources
	*remoteResources = newRemoteResources
	return nil
}

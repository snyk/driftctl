package middlewares

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// The role of this middleware is to expand policy contained in `managed_policy_arns` to dedicated `aws_iam_policy_attachment`
// resources. Note that we do not use `aws_iam_role_policy_attachment` or `aws_iam_user_policy_attachment`
// Once theses resources created, we remove the old `managed_policy_arns` field to avoid false positive drifts

type AwsRoleManagedPolicyExpander struct {
	resourceFactory resource.ResourceFactory
}

func NewAwsRoleManagedPolicyExpander(resourceFactory resource.ResourceFactory) *AwsRoleManagedPolicyExpander {
	return &AwsRoleManagedPolicyExpander{resourceFactory: resourceFactory}
}

func (a AwsRoleManagedPolicyExpander) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newList := make([]*resource.Resource, 0)
	for _, res := range *remoteResources {
		// Ignore all resources other than iam_role
		if res.ResourceType() != aws.AwsIamRoleResourceType {
			newList = append(newList, res)
			continue
		}

		res.Attributes().SafeDelete([]string{"managed_policy_arns"})
		newList = append(newList, res)
	}
	*remoteResources = newList

	newList = make([]*resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than iam_role
		if res.ResourceType() != aws.AwsIamRoleResourceType {
			newList = append(newList, res)
			continue
		}
		managedPolicyArns := res.Attributes().GetSlice("managed_policy_arns")

		// if managed_policy_arns does not exist or is empty ignore resource
		if managedPolicyArns == nil {
			newList = append(newList, res)
			continue
		}

		// Remove empty slices to match remote read results
		if len(managedPolicyArns) == 0 {
			res.Attributes().SafeDelete([]string{"managed_policy_arns"})
			newList = append(newList, res)
			continue
		}

		roleName := res.Attributes().GetString("name")

		for _, arn := range managedPolicyArns {
			arn := arn.(string)
			id := fmt.Sprintf("%s-%s", *roleName, arn)

			policyAttachmentData := resource.Attributes{
				"policy_arn": arn,
				"users":      []interface{}{},
				"groups":     []interface{}{},
				"roles":      []interface{}{*roleName},
			}

			logrus.WithFields(logrus.Fields{
				"role":       *roleName,
				"policy_arn": arn,
			}).Debug("Expanded managed_policy_arns from role")

			newRes := a.resourceFactory.CreateAbstractResource(aws.AwsIamPolicyAttachmentResourceType, id, policyAttachmentData)

			alreadyExist := false
			for _, resInState := range *resourcesFromState {
				if resInState.Equal(newRes) {
					alreadyExist = true
					break
				}
			}

			if !alreadyExist {
				newList = append(newList, newRes)
			}
		}

		res.Attributes().SafeDelete([]string{"managed_policy_arns"})

		newList = append(newList, res)

	}
	*resourcesFromState = newList
	return nil
}

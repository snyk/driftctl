package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// SQS queues from AWS have a weird behaviour when we fetch them.
// By default they have a Policy attached with only an ID
// "arn:aws:sqs:eu-west-3:XXXXXXXXXXXX:foobar/SQSDefaultPolicy" but on fetch
// the SDK return an empty policy (e.g. policy = "").
// We need to ignore those policy from unmanaged resources if they are not managed
// by IaC.
type AwsDefaultSQSQueuePolicy struct{}

func NewAwsDefaultSQSQueuePolicy() AwsDefaultSQSQueuePolicy {
	return AwsDefaultSQSQueuePolicy{}
}

func (m AwsDefaultSQSQueuePolicy) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	newRemoteResources := make([]resource.Resource, 0)
	for _, res := range *remoteResources {
		// Ignore all resources other than sqs_queue_policy
		if res.TerraformType() != aws.AwsSqsQueuePolicyResourceType {
			newRemoteResources = append(newRemoteResources, res)
			continue
		}

		policyRes, _ := res.(*resource.AbstractResource)

		// Ignore all non-default queue policy
		pol, exists := policyRes.Attrs.Get("policy")
		policy := pol.(string)
		if exists && policy != "" {
			newRemoteResources = append(newRemoteResources, policyRes)
			continue
		}

		// Check if queue policy is managed by IaC
		existInState := false
		for _, stateResource := range *resourcesFromState {
			if resource.IsSameResource(res, stateResource) {
				existInState = true
				break
			}
		}

		// Include resource if it's managed in IaC
		if existInState {
			newRemoteResources = append(newRemoteResources, res)
			continue
		}

		// Else, resource is not added to newRemoteResources slice so it will be ignored
		logrus.WithFields(logrus.Fields{
			"id":   res.TerraformId(),
			"type": res.TerraformType(),
		}).Debug("Ignoring default queue policy as it is not managed by IaC")
	}
	*remoteResources = newRemoteResources
	return nil
}

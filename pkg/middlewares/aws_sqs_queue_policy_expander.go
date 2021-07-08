package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// Explodes policy found in aws_sqs_queue.policy from state resources to dedicated resources
type AwsSQSQueuePolicyExpander struct {
	resourceFactory          resource.ResourceFactory
	resourceSchemaRepository resource.SchemaRepositoryInterface
}

func NewAwsSQSQueuePolicyExpander(resourceFactory resource.ResourceFactory, resourceSchemaRepository resource.SchemaRepositoryInterface) AwsSQSQueuePolicyExpander {
	return AwsSQSQueuePolicyExpander{
		resourceFactory,
		resourceSchemaRepository,
	}
}

func (m AwsSQSQueuePolicyExpander) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {
	for _, res := range *remoteResources {
		if res.TerraformType() != aws.AwsSQSQueueResourceType {
			continue
		}
		queue, _ := res.(*resource.AbstractResource)
		queue.Attrs.SafeDelete([]string{"policy"})
	}

	newList := make([]resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than sqs_queue
		if res.TerraformType() != aws.AwsSQSQueueResourceType {
			newList = append(newList, res)
			continue
		}

		queue, _ := res.(*resource.AbstractResource)
		newList = append(newList, res)

		policy, exist := queue.Attrs.Get("policy")
		if !exist || policy == nil {
			continue
		}

		if m.hasPolicyAttached(queue, resourcesFromState) {
			queue.Attrs.SafeDelete([]string{"policy"})
			continue
		}

		err := m.handlePolicy(queue, &newList)
		if err != nil {
			return err
		}
	}
	*resourcesFromState = newList
	return nil
}

func (m *AwsSQSQueuePolicyExpander) handlePolicy(queue *resource.AbstractResource, results *[]resource.Resource) error {
	policy, exists := queue.Attrs.Get("policy")
	if !exists || policy.(string) == "" {
		queue.Attrs.SafeDelete([]string{"policy"})
		return nil
	}

	data := map[string]interface{}{
		"queue_url": queue.Id,
		"id":        queue.Id,
		"policy":    policy,
	}

	newPolicy := m.resourceFactory.CreateAbstractResource("aws_sqs_queue_policy", queue.Id, data)
	*results = append(*results, newPolicy)
	logrus.WithFields(logrus.Fields{
		"id": newPolicy.TerraformId(),
	}).Debug("Created new policy from sqs queue")

	queue.Attrs.SafeDelete([]string{"policy"})
	return nil
}

// Return true if the sqs queue has a aws_sqs_queue_policy resource attached to itself.
// It is mandatory since it's possible to have a aws_sqs_queue with an inline policy
// AND a aws_sqs_queue_policy resource at the same time. At the end, on the AWS console,
// the aws_sqs_queue_policy will be used.
func (m *AwsSQSQueuePolicyExpander) hasPolicyAttached(queue *resource.AbstractResource, resourcesFromState *[]resource.Resource) bool {
	for _, res := range *resourcesFromState {
		if res.TerraformType() == aws.AwsSQSQueuePolicyResourceType &&
			res.TerraformId() == queue.Id {
			return true
		}
	}
	return false
}

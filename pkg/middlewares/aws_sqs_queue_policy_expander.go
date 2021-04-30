package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// Explodes policy found in aws_sqs_queue.policy from state resources to dedicated resources
type AwsSqsQueuePolicyExpander struct {
	resourceFactory          resource.ResourceFactory
	resourceSchemaRepository resource.SchemaRepositoryInterface
}

func NewAwsSqsQueuePolicyExpander(resourceFactory resource.ResourceFactory, resourceSchemaRepository resource.SchemaRepositoryInterface) AwsSqsQueuePolicyExpander {
	return AwsSqsQueuePolicyExpander{
		resourceFactory,
		resourceSchemaRepository,
	}
}

func (m AwsSqsQueuePolicyExpander) Execute(_, resourcesFromState *[]resource.Resource) error {
	newList := make([]resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than sqs_queue
		if res.TerraformType() != aws.AwsSqsQueueResourceType {
			newList = append(newList, res)
			continue
		}

		queue, _ := res.(*aws.AwsSqsQueue)
		newList = append(newList, res)

		if queue.Policy == nil {
			continue
		}

		if m.hasPolicyAttached(queue, resourcesFromState) {
			queue.Policy = nil
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

func (m *AwsSqsQueuePolicyExpander) handlePolicy(queue *aws.AwsSqsQueue, results *[]resource.Resource) error {
	data := map[string]interface{}{
		"queue_url": queue.Id,
		"id":        queue.Id,
		"policy":    queue.Policy,
	}
	ctyVal, err := m.resourceFactory.CreateResource(data, "aws_sqs_queue_policy")
	if err != nil {
		return err
	}

	schema, exist := m.resourceSchemaRepository.GetSchema("aws_ebs_volume")
	ctyAttr := resource.ToResourceAttributes(ctyVal)
	ctyAttr.SanitizeDefaultsV3()
	if exist && schema.NormalizeFunc != nil {
		schema.NormalizeFunc(ctyAttr)
	}

	newPolicy := &resource.AbstractResource{
		Id:    queue.Id,
		Type:  aws.AwsSqsQueuePolicyResourceType,
		Attrs: ctyAttr,
	}

	*results = append(*results, newPolicy)
	logrus.WithFields(logrus.Fields{
		"id": newPolicy.TerraformId(),
	}).Debug("Created new policy from sqs queue")

	queue.Policy = nil
	return nil
}

// Return true if the sqs queue has a aws_sqs_queue_policy resource attached to itself.
// It is mandatory since it's possible to have a aws_sqs_queue with an inline policy
// AND a aws_sqs_queue_policy resource at the same time. At the end, on the AWS console,
// the aws_sqs_queue_policy will be used.
func (m *AwsSqsQueuePolicyExpander) hasPolicyAttached(queue *aws.AwsSqsQueue, resourcesFromState *[]resource.Resource) bool {
	for _, res := range *resourcesFromState {
		if res.TerraformType() == aws.AwsSqsQueuePolicyResourceType &&
			res.TerraformId() == queue.Id {
			return true
		}
	}
	return false
}

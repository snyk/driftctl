package middlewares

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// Explodes policy found in aws_sns_topic from state resources to aws_sns_topic_policy resources
type AwsSNSTopicPolicyExpander struct {
	resourceFactory          resource.ResourceFactory
	resourceSchemaRepository resource.SchemaRepositoryInterface
}

func NewAwsSNSTopicPolicyExpander(resourceFactory resource.ResourceFactory, resourceSchemaRepository resource.SchemaRepositoryInterface) AwsSNSTopicPolicyExpander {
	return AwsSNSTopicPolicyExpander{
		resourceFactory,
		resourceSchemaRepository,
	}
}

func (m AwsSNSTopicPolicyExpander) Execute(_, resourcesFromState *[]resource.Resource) error {
	newList := make([]resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than sns_topic
		if res.TerraformType() != aws.AwsSnsTopicResourceType {
			newList = append(newList, res)
			continue
		}

		topic, _ := res.(*resource.AbstractResource)
		newList = append(newList, res)

		if m.hasPolicyAttached(topic, resourcesFromState) {
			topic.Attrs.SafeDelete([]string{"policy"})
			continue
		}

		err := m.splitPolicy(topic, &newList)
		if err != nil {
			return err
		}
	}
	*resourcesFromState = newList
	return nil
}

func (m *AwsSNSTopicPolicyExpander) splitPolicy(topic *resource.AbstractResource, results *[]resource.Resource) error {
	policy, exist := topic.Attrs.Get("policy")
	if !exist || policy == "" {
		return nil
	}

	arn, exist := topic.Attrs.Get("arn")
	if !exist || arn == "" {
		return errors.Errorf("No arn found for resource %s (%s)", topic.Id, topic.Type)
	}

	data := map[string]interface{}{
		"arn":    arn,
		"id":     topic.Id,
		"policy": policy,
	}

	newPolicy := m.resourceFactory.CreateAbstractResource(topic.Id, "aws_sns_topic_policy", data)

	*results = append(*results, newPolicy)
	logrus.WithFields(logrus.Fields{
		"id": newPolicy.TerraformId(),
	}).Debug("Created new policy from sns_topic")

	topic.Attrs.SafeDelete([]string{"policy"})
	return nil
}

func (m *AwsSNSTopicPolicyExpander) hasPolicyAttached(topic *resource.AbstractResource, resourcesFromState *[]resource.Resource) bool {
	for _, res := range *resourcesFromState {
		if res.TerraformType() == aws.AwsSnsTopicPolicyResourceType &&
			res.TerraformId() == topic.Id {
			return true
		}
	}
	return false
}

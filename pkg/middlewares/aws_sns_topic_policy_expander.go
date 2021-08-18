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

func (m AwsSNSTopicPolicyExpander) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	for _, res := range *remoteResources {
		if res.ResourceType() != aws.AwsSnsTopicResourceType {
			continue
		}
		res.Attrs.SafeDelete([]string{"policy"})
	}

	newList := make([]*resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than sns_topic
		if res.ResourceType() != aws.AwsSnsTopicResourceType {
			newList = append(newList, res)
			continue
		}

		newList = append(newList, res)

		if m.hasPolicyAttached(res, resourcesFromState) {
			res.Attrs.SafeDelete([]string{"policy"})
			continue
		}

		err := m.splitPolicy(res, &newList)
		if err != nil {
			return err
		}
	}
	*resourcesFromState = newList
	return nil
}

func (m *AwsSNSTopicPolicyExpander) splitPolicy(topic *resource.Resource, results *[]*resource.Resource) error {
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

	newPolicy := m.resourceFactory.CreateAbstractResource("aws_sns_topic_policy", topic.Id, data)

	*results = append(*results, newPolicy)
	logrus.WithFields(logrus.Fields{
		"id": newPolicy.ResourceId(),
	}).Debug("Created new policy from sns_topic")

	topic.Attrs.SafeDelete([]string{"policy"})
	return nil
}

func (m *AwsSNSTopicPolicyExpander) hasPolicyAttached(topic *resource.Resource, resourcesFromState *[]*resource.Resource) bool {
	for _, res := range *resourcesFromState {
		if res.ResourceType() == aws.AwsSnsTopicPolicyResourceType &&
			res.ResourceId() == topic.Id {
			return true
		}
	}
	return false
}

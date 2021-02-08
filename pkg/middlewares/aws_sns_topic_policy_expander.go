package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// Explodes policy found in aws_sns_topic from state resources to aws_sns_topic_policy resources
type AwsSNSTopicPolicyExpander struct{}

func NewAwsSNSTopicPolicyExpander() AwsSNSTopicPolicyExpander {
	return AwsSNSTopicPolicyExpander{}
}

func (m AwsSNSTopicPolicyExpander) Execute(_, resourcesFromState *[]resource.Resource) error {
	newList := make([]resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than sns_topic
		if res.TerraformType() != aws.AwsSnsTopicResourceType {
			newList = append(newList, res)
			continue
		}

		topic, _ := res.(*aws.AwsSnsTopic)
		newList = append(newList, res)

		if m.hasPolicyAttached(topic, resourcesFromState) {
			topic.Policy = nil
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

func (m *AwsSNSTopicPolicyExpander) splitPolicy(topic *aws.AwsSnsTopic, results *[]resource.Resource) error {
	if topic.Policy == nil || *topic.Policy == "" {
		return nil
	}

	newPolicy := &aws.AwsSnsTopicPolicy{
		Id:     topic.Id,
		Arn:    topic.Arn,
		Policy: topic.Policy,
	}

	normalized, err := newPolicy.NormalizeForState()
	if err != nil {
		return err
	}

	*results = append(*results, normalized)
	logrus.WithFields(logrus.Fields{
		"id": newPolicy.TerraformId(),
	}).Debug("Created new policy from sns_topic")

	topic.Policy = nil
	return nil
}

func (m *AwsSNSTopicPolicyExpander) hasPolicyAttached(topic *aws.AwsSnsTopic, resourcesFromState *[]resource.Resource) bool {
	for _, res := range *resourcesFromState {
		if res.TerraformType() == aws.AwsSnsTopicPolicyResourceType &&
			res.TerraformId() == topic.Id {
			return true
		}
	}
	return false
}
